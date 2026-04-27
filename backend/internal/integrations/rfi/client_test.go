package rfi

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewClientNotConfigured(t *testing.T) {
	c, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Configured() {
		t.Fatalf("expected not configured")
	}
	_, err = c.ListAvailableSlots(context.Background(), "VR", "MU", time.Now())
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestClientConfigured(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slots":[{"id":"slot-1","terminal":"VR-QE","status":"confirmed"}]}`))
	}))
	defer srv.Close()
	c, err := New(Config{BaseURL: srv.URL, ClientID: "id", ClientSecret: "s"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	slots, err := c.ListAvailableSlots(context.Background(), "VR", "MU", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 1 || slots[0].ID != "slot-1" {
		t.Fatalf("unexpected payload: %+v", slots)
	}
}

// TestNewMTLSConfigIncomplete: providing only MTLSCertFile (or only
// MTLSKeyFile) is a misconfiguration. Both must be present, or
// neither. The previous code silently ignored both fields, which
// hid this exact misconfiguration.
func TestNewMTLSConfigIncomplete(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"cert_without_key", Config{MTLSCertFile: "/tmp/cert.pem"}},
		{"key_without_cert", Config{MTLSKeyFile: "/tmp/key.pem"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.cfg)
			if !errors.Is(err, ErrMTLSConfigIncomplete) {
				t.Fatalf("want ErrMTLSConfigIncomplete, got %v", err)
			}
		})
	}
}

// TestMTLSHandshake spins up an httptest TLS server that *requires*
// a client certificate, then proves:
//
//   1. A client built without mTLS cannot complete the handshake.
//   2. A client built with the matching mTLS keypair completes the
//      handshake and gets a 200 from the server.
//   3. UsesMTLS() returns true when the cert pair is loaded, false
//      otherwise.
//
// The test generates a self-signed CA, a server cert signed by it,
// and a client cert signed by it — all on the fly. No filesystem
// fixtures committed to the repo.
func TestMTLSHandshake(t *testing.T) {
	caPEM, caKey := newCA(t)
	clientCertPEM, clientKeyPEM := signedLeaf(t, caPEM, caKey, "logitrack-test-client")
	serverCertPEM, serverKeyPEM := signedLeaf(t, caPEM, caKey, "127.0.0.1")

	dir := t.TempDir()
	caFile := writeFile(t, dir, "ca.pem", caPEM)
	clientCertFile := writeFile(t, dir, "client-cert.pem", clientCertPEM)
	clientKeyFile := writeFile(t, dir, "client-key.pem", clientKeyPEM)

	// Server: requires client certs verified against our CA.
	serverPair, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatalf("load server cert: %v", err)
	}
	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(caPEM) {
		t.Fatal("append CA")
	}
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slots":[{"id":"mtls-ok","terminal":"VR-QE","status":"confirmed"}]}`))
	}))
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverPair},
		ClientCAs:    clientCAs,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}
	srv.StartTLS()
	defer srv.Close()

	t.Run("without_mtls_handshake_fails", func(t *testing.T) {
		// Build a client that trusts the server CA but has no client cert.
		c, err := New(Config{
			BaseURL:      srv.URL,
			ClientID:     "id",
			ClientSecret: "s",
			MTLSCAFile:   caFile,
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if c.UsesMTLS() {
			t.Fatal("expected UsesMTLS=false without cert pair")
		}
		_, err = c.ListAvailableSlots(context.Background(), "VR", "MU", time.Now())
		if err == nil {
			t.Fatal("expected handshake failure without client cert")
		}
	})

	t.Run("with_mtls_handshake_succeeds", func(t *testing.T) {
		c, err := New(Config{
			BaseURL:      srv.URL,
			ClientID:     "id",
			ClientSecret: "s",
			MTLSCertFile: clientCertFile,
			MTLSKeyFile:  clientKeyFile,
			MTLSCAFile:   caFile,
		})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if !c.UsesMTLS() {
			t.Fatal("expected UsesMTLS=true with cert pair")
		}
		slots, err := c.ListAvailableSlots(context.Background(), "VR", "MU", time.Now())
		if err != nil {
			t.Fatalf("expected success with mTLS, got %v", err)
		}
		if len(slots) != 1 || slots[0].ID != "mtls-ok" {
			t.Fatalf("unexpected payload: %+v", slots)
		}
	})
}

// --- test helpers ----------------------------------------------------------

// newCA returns (caCertPEM, caPrivateKey) for a freshly-generated
// self-signed CA usable to sign leaf certs in this test only.
func newCA(t *testing.T) ([]byte, *ecdsa.PrivateKey) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ca key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "logitrack-test-ca"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(2 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, priv.Public(), priv)
	if err != nil {
		t.Fatalf("ca create: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), priv
}

// signedLeaf returns (certPEM, keyPEM) for a leaf cert signed by the
// supplied CA. CommonName ends up in DNS SAN for "127.0.0.1" / hostname
// matching during the TLS handshake.
func signedLeaf(t *testing.T, caPEM []byte, caKey *ecdsa.PrivateKey, commonName string) ([]byte, []byte) {
	t.Helper()
	caBlock, _ := pem.Decode(caPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		t.Fatalf("parse ca: %v", err)
	}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("leaf key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(2 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     []string{commonName, "localhost"},
	}
	if commonName == "127.0.0.1" {
		tmpl.IPAddresses = append(tmpl.IPAddresses, parseIPv4("127.0.0.1"))
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, caCert, priv.Public(), caKey)
	if err != nil {
		t.Fatalf("leaf create: %v", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM
}

func writeFile(t *testing.T, dir, name string, body []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, body, 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

// parseIPv4 is a small helper to avoid pulling net into the test set
// just for one constant. The crypto/x509 IPAddresses field expects
// net.IP, which is []byte under the hood.
func parseIPv4(s string) []byte {
	var out [4]byte
	var n int
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			n++
			continue
		}
		out[n] = out[n]*10 + s[i] - '0'
	}
	return out[:]
}
