package services

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/logitrack/backend/internal/config"
	"github.com/logitrack/backend/internal/modules/logistics"
)

// RouteOptimizer encapsulates the OSRM integration used for route
// planning, ETA calculation and multi-stop optimisation. With more
// than two waypoints the OSRM trip service reorders the intermediate
// stops (first and last stay pinned) and the optimised visiting order
// is reported in RouteResponse.WaypointOrder; with exactly two the
// route service is queried as-is. The service is exposed via an
// interface so unit tests can substitute a mock without a live HTTP
// backend.
type RouteOptimizer interface {
	OptimiseRoute(ctx context.Context, req RouteRequest) (*RouteResponse, error)
	EstimateETA(ctx context.Context, from, to logistics.GeoPoint) (time.Duration, error)
}

// RouteRequest is the input to OptimiseRoute. Vehicle is one of
// "truck", "van" or "car"; AvoidTolls requests the alternate profile.
type RouteRequest struct {
	Waypoints  []logistics.GeoPoint `json:"waypoints"`
	Vehicle    string               `json:"vehicle"`
	AvoidTolls bool                 `json:"avoidTolls"`
}

// RouteResponse is the normalised result returned to callers. The
// geometry is Polyline6 encoded (OSRM default) so the frontend can
// decode and draw it on Leaflet/MapLibre. Source reports whether the
// answer came from OSRM, the in-memory cache, or the local fallback.
type RouteResponse struct {
	Distance float64 `json:"distanceMeters"`
	Duration float64 `json:"durationSeconds"`
	Geometry string  `json:"geometry"`
	Legs     []struct {
		Distance float64 `json:"distanceMeters"`
		Duration float64 `json:"durationSeconds"`
	} `json:"legs"`
	// WaypointOrder lists the indices of the request waypoints in
	// visiting order. The trip service may permute the intermediate
	// stops; route and fallback answers report the identity order.
	WaypointOrder []int  `json:"waypointOrder,omitempty"`
	Source        string `json:"source,omitempty"` // "osrm" | "cache" | "fallback"
	Notes         string `json:"notes,omitempty"`
}

// DefaultAverageSpeedKPH is the conservative driving speed used by the
// fallback estimator when OSRM is unreachable. 70 km/h matches the
// observed average on the A4/A22 Verona–Milano/Verona–Brennero
// corridors accounting for urban legs and rest stops.
const DefaultAverageSpeedKPH = 70.0

// ErrOSRMHostNotAllowed is returned when the configured BaseURL resolves
// to a host not on OSRMConfig.AllowedHosts.
var ErrOSRMHostNotAllowed = errors.New("osrm: host not in allow-list")

// OSRMOptimizer is the default production implementation.
type OSRMOptimizer struct {
	cfg                config.OSRMConfig
	client             *http.Client
	log                *zap.Logger
	cache              *lruCache
	truckDowngradeWarn sync.Once
	noOSRMWarn         sync.Once
}

// NewOSRMOptimizer wires a RouteOptimizer backed by an OSRM server.
// The returned implementation is safe for concurrent use.
//
// The HTTP client refuses to follow redirects. A compromised or
// MITM'd OSRM service could otherwise return a 30x to an internal
// host (e.g. 169.254.169.254 metadata) and bypass the BaseURL
// allow-list, since the allow-list is only checked for the initial
// request URL.
func NewOSRMOptimizer(cfg config.OSRMConfig, log *zap.Logger) *OSRMOptimizer {
	size := cfg.CacheSize
	if size <= 0 {
		size = 1000
	}
	return &OSRMOptimizer{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		log:   log,
		cache: newLRUCache(size),
	}
}

// OptimiseRoute resolves the requested waypoints against OSRM with
// three layers:
//
//  1. In-memory LRU cache keyed by the canonical coordinate string and
//     the OSRM service the request maps to (trip vs route).
//  2. OSRM HTTP call to the allow-listed BaseURL: /trip/v1 with
//     roundtrip=false&source=first&destination=last when there are
//     more than two waypoints, /route/v1 in the given order otherwise.
//     A /trip failure retries as /route before giving up on OSRM.
//  3. Straight-line distance × DefaultAverageSpeedKPH fallback if OSRM
//     is unreachable (network error, timeout, non-2xx).
func (o *OSRMOptimizer) OptimiseRoute(ctx context.Context, req RouteRequest) (*RouteResponse, error) {
	if len(req.Waypoints) < 2 {
		return nil, fmt.Errorf("route requires >= 2 waypoints, got %d", len(req.Waypoints))
	}
	for _, wp := range req.Waypoints {
		if len(wp.Coordinates) != 2 {
			return nil, logistics.ErrInvalidGeoPoint
		}
	}

	key := cacheKey(req)
	if hit, ok := o.cache.get(key); ok {
		clone := *hit
		clone.Source = "cache"
		return &clone, nil
	}

	// No OSRM configured = fall back silently after a one-shot WARN.
	// This is the kit default: OSRM is opt-in, customers wire their own
	// self-hosted instance per deployment.
	if strings.TrimSpace(o.cfg.BaseURL) == "" {
		o.noOSRMWarn.Do(func() {
			o.log.Warn("osrm: OSRM_BASE_URL not configured, using straight-line fallback for every route request")
		})
		return o.fallback(req), nil
	}
	// Validate the configured host against the allow-list BEFORE any
	// DNS resolution or outbound socket is opened. SSRF mitigation:
	// even with write access to OSRM_BASE_URL an attacker cannot
	// redirect traffic to an unexpected host.
	base, err := url.Parse(strings.TrimRight(o.cfg.BaseURL, "/"))
	if err != nil || base.Host == "" {
		o.log.Warn("osrm: invalid base url, using fallback", zap.Error(err))
		return o.fallback(req), nil
	}
	if !o.hostAllowed(base.Hostname()) {
		o.log.Warn("osrm: host not in allow-list", zap.String("host", base.Hostname()))
		return nil, ErrOSRMHostNotAllowed
	}

	osrmResp, err := o.callOSRM(ctx, base, req)
	if err != nil {
		o.log.Warn("osrm: call failed, using fallback", zap.Error(err))
		fb := o.fallback(req)
		o.cache.put(key, fb) // cache fallback briefly
		return fb, nil
	}
	osrmResp.Source = "osrm"
	o.cache.put(key, osrmResp)
	return osrmResp, nil
}

// EstimateETA is a convenience wrapper returning only the duration.
func (o *OSRMOptimizer) EstimateETA(ctx context.Context, from, to logistics.GeoPoint) (time.Duration, error) {
	r, err := o.OptimiseRoute(ctx, RouteRequest{Waypoints: []logistics.GeoPoint{from, to}})
	if err != nil {
		return 0, err
	}
	return time.Duration(r.Duration * float64(time.Second)), nil
}

// hostAllowed checks the configured allow-list. Hosts are compared
// case-insensitively; no wildcard support — a deliberate choice to
// keep the surface area tight.
func (o *OSRMOptimizer) hostAllowed(host string) bool {
	h := strings.ToLower(host)
	for _, allowed := range o.cfg.AllowedHosts {
		if strings.EqualFold(strings.TrimSpace(allowed), h) {
			return true
		}
	}
	// As an additional guard, refuse any literal IP address that is not
	// also listed by address (no metadata services, no loopback tricks).
	if ip := net.ParseIP(host); ip != nil {
		for _, allowed := range o.cfg.AllowedHosts {
			if strings.EqualFold(strings.TrimSpace(allowed), ip.String()) {
				return true
			}
		}
		return false
	}
	return false
}

// callOSRM picks the OSRM service for the request: /trip/v1 when there
// are intermediate stops to reorder, /route/v1 for a plain A→B pair.
// A /trip failure degrades to /route in the given order before the
// caller resorts to the haversine fallback, so a reachable OSRM
// without trip support still yields real road distances.
func (o *OSRMOptimizer) callOSRM(ctx context.Context, base *url.URL, req RouteRequest) (*RouteResponse, error) {
	if len(req.Waypoints) > 2 {
		resp, err := o.callTrip(ctx, base, req)
		if err == nil {
			return resp, nil
		}
		o.log.Warn("osrm: trip failed, retrying as route in the given order", zap.Error(err))
	}
	return o.callRoute(ctx, base, req)
}

// callTrip solves the visiting order via the OSRM trip service. The
// first and last waypoints stay pinned (roundtrip=false, source=first,
// destination=last): departure and final delivery are fixed by the
// shipment, only the intermediate stops are up for reordering.
func (o *OSRMOptimizer) callTrip(ctx context.Context, base *url.URL, req RouteRequest) (*RouteResponse, error) {
	ref := *base
	ref.Path = fmt.Sprintf("/trip/v1/%s/%s", o.profileFor(req), coordinatePath(req.Waypoints))
	q := url.Values{}
	q.Set("overview", "full")
	q.Set("geometries", "polyline6")
	q.Set("steps", "false")
	q.Set("roundtrip", "false")
	q.Set("source", "first")
	q.Set("destination", "last")
	if req.AvoidTolls {
		q.Set("exclude", "toll")
	}
	ref.RawQuery = q.Encode()

	var decoded struct {
		Code      string `json:"code"`
		Waypoints []struct {
			WaypointIndex int `json:"waypoint_index"`
		} `json:"waypoints"`
		Trips []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry string  `json:"geometry"`
			Legs     []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
			} `json:"legs"`
		} `json:"trips"`
	}
	if err := o.getJSON(ctx, ref.String(), &decoded); err != nil {
		return nil, err
	}
	if decoded.Code != "Ok" || len(decoded.Trips) == 0 {
		return nil, fmt.Errorf("osrm no trip: %s", decoded.Code)
	}
	if len(decoded.Waypoints) != len(req.Waypoints) {
		return nil, fmt.Errorf("osrm trip returned %d waypoints for %d requested", len(decoded.Waypoints), len(req.Waypoints))
	}
	// waypoint_index is the position of each input waypoint within the
	// trip; inverting it yields the visiting order. A malformed index
	// set aborts the trip answer instead of shipping a bogus order.
	order := make([]int, len(decoded.Waypoints))
	seen := make([]bool, len(decoded.Waypoints))
	for inputIdx, wp := range decoded.Waypoints {
		if wp.WaypointIndex < 0 || wp.WaypointIndex >= len(order) || seen[wp.WaypointIndex] {
			return nil, fmt.Errorf("osrm trip: invalid waypoint_index %d", wp.WaypointIndex)
		}
		seen[wp.WaypointIndex] = true
		order[wp.WaypointIndex] = inputIdx
	}
	trip := decoded.Trips[0]
	out := &RouteResponse{Distance: trip.Distance, Duration: trip.Duration, Geometry: trip.Geometry, WaypointOrder: order}
	for _, leg := range trip.Legs {
		out.Legs = append(out.Legs, struct {
			Distance float64 `json:"distanceMeters"`
			Duration float64 `json:"durationSeconds"`
		}{leg.Distance, leg.Duration})
	}
	return out, nil
}

// callRoute queries the OSRM route service with the waypoints in the
// given order.
func (o *OSRMOptimizer) callRoute(ctx context.Context, base *url.URL, req RouteRequest) (*RouteResponse, error) {
	ref := *base
	ref.Path = fmt.Sprintf("/route/v1/%s/%s", o.profileFor(req), coordinatePath(req.Waypoints))
	q := url.Values{}
	q.Set("overview", "full")
	q.Set("geometries", "polyline6")
	q.Set("alternatives", "false")
	q.Set("steps", "false")
	if req.AvoidTolls {
		q.Set("exclude", "toll")
	}
	ref.RawQuery = q.Encode()

	var decoded struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry string  `json:"geometry"`
			Legs     []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
			} `json:"legs"`
		} `json:"routes"`
	}
	if err := o.getJSON(ctx, ref.String(), &decoded); err != nil {
		return nil, err
	}
	if decoded.Code != "Ok" || len(decoded.Routes) == 0 {
		return nil, fmt.Errorf("osrm no route: %s", decoded.Code)
	}
	r := decoded.Routes[0]
	out := &RouteResponse{Distance: r.Distance, Duration: r.Duration, Geometry: r.Geometry, WaypointOrder: identityOrder(len(req.Waypoints))}
	for _, leg := range r.Legs {
		out.Legs = append(out.Legs, struct {
			Distance float64 `json:"distanceMeters"`
			Duration float64 `json:"durationSeconds"`
		}{leg.Distance, leg.Duration})
	}
	return out, nil
}

// profileFor maps the requested vehicle onto an OSRM profile name.
func (o *OSRMOptimizer) profileFor(req RouteRequest) string {
	profile := "driving"
	if req.Vehicle == "truck" {
		// Production deployments serve a custom truck profile (e.g.
		// "truck" or "hgv") with weight, height, hazardous-goods and
		// ZTL awareness. The public OSRM service ships only "driving",
		// and the truck profile is opt-in via OSRM_TRUCK_PROFILE.
		if o.cfg.TruckProfile != "" {
			profile = o.cfg.TruckProfile
		} else {
			// One-shot WARN per process: silently downgrading every
			// truck request would hide the fact that the route plan
			// ignores HGV restrictions.
			o.truckDowngradeWarn.Do(func() {
				o.log.Warn("osrm: vehicle=truck requested but OSRM_TRUCK_PROFILE is not set; falling back to driving profile (HGV restrictions, weight limits, hazardous-goods rules will be ignored). Configure a custom OSRM truck profile to silence this.",
					zap.String("base_url", o.cfg.BaseURL),
				)
			})
		}
	}
	return profile
}

// getJSON issues the OSRM GET and decodes the 200 body into out.
func (o *OSRMOptimizer) getJSON(ctx context.Context, rawURL string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("User-Agent", "logitrack-backend/1.0 (+https://logitrack.it)")
	resp, err := o.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("osrm get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("osrm status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("osrm decode: %w", err)
	}
	return nil
}

// coordinatePath renders the waypoints in the {lon},{lat};… form OSRM
// expects in the request path.
func coordinatePath(wps []logistics.GeoPoint) string {
	parts := make([]string, 0, len(wps))
	for _, wp := range wps {
		parts = append(parts, fmt.Sprintf("%.6f,%.6f", wp.Coordinates[0], wp.Coordinates[1]))
	}
	return strings.Join(parts, ";")
}

// identityOrder is the visiting order when no reordering happened.
func identityOrder(n int) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	return order
}

// fallback produces a conservative estimate by summing the great-circle
// distances between successive waypoints and assuming DefaultAverageSpeedKPH.
// The geometry is left empty so the frontend can fall back to a
// straight-line polyline between waypoints.
func (o *OSRMOptimizer) fallback(req RouteRequest) *RouteResponse {
	totalMeters := 0.0
	for i := 1; i < len(req.Waypoints); i++ {
		a := req.Waypoints[i-1].Coordinates
		b := req.Waypoints[i].Coordinates
		km := haversineKm(a[1], a[0], b[1], b[0])
		totalMeters += km * 1000
	}
	seconds := (totalMeters / 1000) / DefaultAverageSpeedKPH * 3600
	return &RouteResponse{
		Distance:      totalMeters,
		Duration:      seconds,
		Geometry:      "",
		WaypointOrder: identityOrder(len(req.Waypoints)),
		Source:        "fallback",
		Notes:         "OSRM unreachable; straight-line estimate at 70 km/h average",
	}
}

// cacheKey builds a compact, deterministic key from the request. The
// OSRM service the request maps to is part of the key: /trip may
// reorder the stops while /route never does, so an answer cached for
// one must not be served for the other.
func cacheKey(req RouteRequest) string {
	var sb strings.Builder
	if len(req.Waypoints) > 2 {
		sb.WriteString("trip|")
	} else {
		sb.WriteString("route|")
	}
	sb.WriteString(req.Vehicle)
	if req.AvoidTolls {
		sb.WriteByte('T')
	}
	for _, wp := range req.Waypoints {
		if len(wp.Coordinates) != 2 {
			continue
		}
		_, _ = fmt.Fprintf(&sb, "|%.5f,%.5f", wp.Coordinates[0], wp.Coordinates[1])
	}
	return sb.String()
}

// --- lruCache --------------------------------------------------------------

// lruCache is a tiny thread-safe LRU tuned for the route-optimizer hot
// path. We do not import a third-party LRU to keep the supply-chain
// surface minimal — the implementation is ~40 lines and covered by a
// dedicated unit test.
type lruCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type lruItem struct {
	key   string
	value *RouteResponse
}

func newLRUCache(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *lruCache) get(key string) (*RouteResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*lruItem).value, true
}

func (c *lruCache) put(key string, value *RouteResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		el.Value.(*lruItem).value = value
		return
	}
	el := c.order.PushFront(&lruItem{key: key, value: value})
	c.items[key] = el
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*lruItem).key)
		}
	}
}

// Size returns the current number of cached entries (test visibility).
func (c *lruCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}
