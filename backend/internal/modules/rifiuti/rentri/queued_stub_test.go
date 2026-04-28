package rentri

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func newDeterministicStub(t *testing.T) *QueuedStub {
	t.Helper()
	frozen := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	return NewQueuedStub().WithClock(func() time.Time { return frozen })
}

func TestQueuedStubVidimaFIRDeterministic(t *testing.T) {
	t.Parallel()
	stub := newDeterministicStub(t)
	ctx := context.Background()
	req := VidimazioneRequest{
		TenantID:       "tenant-FRO",
		IdempotencyKey: "FIR-2026-0001",
		XFIRPayload:    []byte(`<formulario><cer>150106</cer></formulario>`),
	}

	first, err := stub.VidimaFIR(ctx, req)
	if err != nil {
		t.Fatalf("first VidimaFIR error: %v", err)
	}
	second, err := stub.VidimaFIR(ctx, req)
	if err != nil {
		t.Fatalf("second VidimaFIR error: %v", err)
	}
	if first.NumeroRENTRI != second.NumeroRENTRI {
		t.Fatalf("idempotent vidimazione should yield identical numero: %q vs %q", first.NumeroRENTRI, second.NumeroRENTRI)
	}
	if first.QRCodePayload != second.QRCodePayload {
		t.Fatalf("QR code payload should be deterministic")
	}
}

func TestQueuedStubRequiresIdempotency(t *testing.T) {
	t.Parallel()
	stub := NewQueuedStub()
	ctx := context.Background()
	_, err := stub.VidimaFIR(ctx, VidimazioneRequest{
		TenantID:    "t",
		XFIRPayload: []byte("x"),
	})
	if !errors.Is(err, ErrIdempotencyRequired) {
		t.Fatalf("expected ErrIdempotencyRequired, got %v", err)
	}
}

func TestQueuedStubRejectsEmptyPayload(t *testing.T) {
	t.Parallel()
	stub := NewQueuedStub()
	ctx := context.Background()
	_, err := stub.VidimaFIR(ctx, VidimazioneRequest{
		TenantID:       "t",
		IdempotencyKey: "k",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestQueuedStubMovimentoAccumulates(t *testing.T) {
	t.Parallel()
	stub := newDeterministicStub(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := stub.TrasmettiMovimento(ctx, MovimentoRequest{
			TenantID:       "t",
			XMLPayload:     []byte("<m/>"),
			IdempotencyKey: "k-" + time.Now().Format("000000000") + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("TrasmettiMovimento %d: %v", i, err)
		}
	}
	pending := stub.Pending()
	if len(pending) != 3 {
		t.Fatalf("expected 3 pending items, got %d", len(pending))
	}
	drained := stub.Drain()
	if len(drained) != 3 {
		t.Fatalf("expected 3 drained, got %d", len(drained))
	}
	if len(stub.Pending()) != 0 {
		t.Fatalf("Drain should empty the queue")
	}
}

func TestQueuedStubGetFIRRoundTrip(t *testing.T) {
	t.Parallel()
	stub := newDeterministicStub(t)
	ctx := context.Background()
	resp, err := stub.VidimaFIR(ctx, VidimazioneRequest{
		TenantID:       "t",
		IdempotencyKey: "k",
		XFIRPayload:    []byte("<f/>"),
	})
	if err != nil {
		t.Fatalf("vidimazione error: %v", err)
	}
	got, err := stub.GetFIR(ctx, resp.NumeroRENTRI)
	if err != nil {
		t.Fatalf("GetFIR error: %v", err)
	}
	if got.NumeroRENTRI != resp.NumeroRENTRI {
		t.Fatalf("GetFIR returned wrong numero: %q", got.NumeroRENTRI)
	}
	if got.State != "vidimato" {
		t.Fatalf("GetFIR returned wrong state: %q", got.State)
	}
}

func TestQueuedStubUnknownNumeroNotFound(t *testing.T) {
	t.Parallel()
	stub := NewQueuedStub()
	if _, err := stub.GetFIR(context.Background(), "missing"); !errors.Is(err, ErrFIRNotFound) {
		t.Fatalf("expected ErrFIRNotFound, got %v", err)
	}
	if _, err := stub.GetFIRXML(context.Background(), "missing"); !errors.Is(err, ErrFIRNotFound) {
		t.Fatalf("expected ErrFIRNotFound, got %v", err)
	}
}

func TestQueuedStubConcurrent(t *testing.T) {
	t.Parallel()
	stub := newDeterministicStub(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	const workers = 16
	const perWorker = 32
	for w := 0; w < workers; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				_, err := stub.TrasmettiMovimento(ctx, MovimentoRequest{
					TenantID:       "t",
					XMLPayload:     []byte("<m/>"),
					IdempotencyKey: time.Now().Format("000000") + string(rune('A'+w)) + string(rune('0'+i%10)),
				})
				if err != nil {
					t.Errorf("worker %d iter %d: %v", w, i, err)
					return
				}
			}
		}()
	}
	wg.Wait()
	if got := len(stub.Pending()); got != workers*perWorker {
		t.Fatalf("expected %d pending entries, got %d", workers*perWorker, got)
	}
}

func TestEnvironmentBaseURL(t *testing.T) {
	t.Parallel()
	if EnvSandbox.BaseURL() != SandboxAPIBase {
		t.Fatalf("sandbox base url mismatch")
	}
	if EnvProduction.BaseURL() != ProductionAPIBase {
		t.Fatalf("production base url mismatch")
	}
	if Environment("anything-else").BaseURL() != SandboxAPIBase {
		t.Fatalf("unknown env should default to sandbox")
	}
}
