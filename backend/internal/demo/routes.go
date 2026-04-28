// Package demo contains deterministic fixtures used by the golden-path
// E2E test and the compose "demo" profile. Nothing under this package
// may be imported by production code paths outside the seed flow.
package demo

import "github.com/logitrack/backend/internal/modules/logistics"

// Waypoint is a (lon, lat) pair used to feed the simulator. The order
// matches GeoJSON convention and the MongoDB 2dsphere index.
type Waypoint struct {
	Lon, Lat float64
}

// Route is a pre-computed demo polyline.
type Route struct {
	ID          string
	Reference   string
	Carrier     string
	OriginName  string
	DestName    string
	Origin      Waypoint
	Destination Waypoint
	Polyline    []Waypoint
	Distance    float64 // metres, pre-computed
}

// --- Routes ----------------------------------------------------------------
// Coordinates are taken from publicly available maps (OpenStreetMap).
// Accuracy is sufficient for a demo: the points trace the main corridor
// but do not claim lane-level precision. Production deployments would
// pull the geometry from the self-hosted OSRM /route endpoint.

// VeronaMilano traces A4 Verona Sud → Milano Est via Brescia Est and
// Bergamo Est.
var VeronaMilano = Route{
	ID:          "demo-shipment-verona-milano",
	Reference:   "CMR-VRMI-001",
	Carrier:     "Autotrasporti del Garda Srl",
	OriginName:  "Mozzecane (VR)",
	DestName:    "Milano Smistamento",
	Origin:      Waypoint{Lon: 10.793, Lat: 45.341}, // Mozzecane
	Destination: Waypoint{Lon: 9.214, Lat: 45.450},  // Milano Lambrate
	Polyline: []Waypoint{
		{Lon: 10.7930, Lat: 45.3410}, // Mozzecane
		{Lon: 10.7220, Lat: 45.3810}, // Verona Sud casello
		{Lon: 10.5000, Lat: 45.4250}, // Peschiera del Garda
		{Lon: 10.3010, Lat: 45.4430}, // Sirmione
		{Lon: 10.2100, Lat: 45.5420}, // Brescia Est casello
		{Lon: 10.0180, Lat: 45.5500}, // Brescia Centro
		{Lon: 9.8500, Lat: 45.6000},  // Rovato
		{Lon: 9.6700, Lat: 45.6700},  // Bergamo Est
		{Lon: 9.5350, Lat: 45.6100},  // Capriate San Gervasio
		{Lon: 9.3800, Lat: 45.5500},  // Trezzo sull'Adda
		{Lon: 9.2800, Lat: 45.5200},  // Cassina de' Pecchi
		{Lon: 9.2140, Lat: 45.4500},  // Milano Lambrate
	},
	Distance: 160_000,
}

// VeronaNapoli traces A1 Verona → Bologna → Firenze → Roma → Napoli.
// Shortened polyline; the simulator interpolates between adjacent
// segments at 1 Hz.
var VeronaNapoli = Route{
	ID:          "demo-shipment-verona-napoli",
	Reference:   "CMR-VRNA-002",
	Carrier:     "Trasporti Mediterraneo Spa",
	OriginName:  "Quadrante Europa (VR)",
	DestName:    "Napoli Est Interporto",
	Origin:      Waypoint{Lon: 10.965, Lat: 45.398},
	Destination: Waypoint{Lon: 14.356, Lat: 40.843},
	Polyline: []Waypoint{
		{Lon: 10.9650, Lat: 45.3980}, // Quadrante Europa
		{Lon: 11.0000, Lat: 45.1500}, // Mantova
		{Lon: 11.3270, Lat: 44.4900}, // Bologna
		{Lon: 11.2500, Lat: 43.7700}, // Firenze Scandicci
		{Lon: 11.2550, Lat: 43.3800}, // Valdichiana
		{Lon: 12.4000, Lat: 42.7000}, // Orte
		{Lon: 12.5000, Lat: 41.9000}, // Roma Nord
		{Lon: 13.0000, Lat: 41.5000}, // Frosinone
		{Lon: 13.7500, Lat: 41.2000}, // Cassino
		{Lon: 14.2500, Lat: 41.0000}, // Caserta Nord
		{Lon: 14.3560, Lat: 40.8430}, // Napoli Est
	},
	Distance: 825_000,
}

// VeronaMunchen traces A22 Brennero to München via Innsbruck.
var VeronaMunchen = Route{
	ID:          "demo-shipment-verona-munchen",
	Reference:   "CMR-VRMU-003",
	Carrier:     "Brennero Logistik GmbH",
	OriginName:  "Mozzecane (VR)",
	DestName:    "München Riem Hub",
	Origin:      Waypoint{Lon: 10.793, Lat: 45.341},
	Destination: Waypoint{Lon: 11.700, Lat: 48.140},
	Polyline: []Waypoint{
		{Lon: 10.7930, Lat: 45.3410}, // Mozzecane
		{Lon: 10.9600, Lat: 45.4000}, // Verona Nord casello
		{Lon: 10.8600, Lat: 45.5200}, // Rovereto Sud
		{Lon: 11.1200, Lat: 46.0700}, // Trento
		{Lon: 11.3370, Lat: 46.4960}, // Bolzano
		{Lon: 11.4450, Lat: 46.7800}, // Bressanone
		{Lon: 11.5100, Lat: 47.0020}, // Brennero / Brenner
		{Lon: 11.4000, Lat: 47.2700}, // Innsbruck
		{Lon: 11.6200, Lat: 47.7000}, // Wiesing / Achensee
		{Lon: 11.7000, Lat: 48.0000}, // Rosenheim Nord
		{Lon: 11.7000, Lat: 48.1400}, // München Riem
	},
	Distance: 412_000,
}

// AllRoutes returns the three demo routes in display order.
func AllRoutes() []Route {
	return []Route{VeronaMilano, VeronaNapoli, VeronaMunchen}
}

// ShipmentTemplate renders a minimally populated Shipment ready for
// persistence. The tenantID is injected by the seeder so the demo is
// scoped to its own tenant.
func (r Route) ShipmentTemplate(tenantID string) *logistics.Shipment {
	return &logistics.Shipment{
		ID:        r.ID,
		TenantID:  tenantID,
		Reference: r.Reference,
		Carrier:   r.Carrier,
		Mode:      logistics.ModeRoad,
		Status:    logistics.StatusInTransit,
		Consignor: logistics.Party{
			Name: "LogiTrack Demo Consignor Srl", City: r.OriginName, Country: "IT",
			VATNumber: "IT01234567890", PostalCode: "37060", Province: "VR", Address: "Via dell'Industria 1",
		},
		Consignee: logistics.Party{
			Name: "LogiTrack Demo Consignee", City: r.DestName, Country: countryOf(r.DestName),
			VATNumber: "IT09876543210", Address: "Via del Terminal 1",
		},
		Origin:       logistics.NewGeoPoint(r.Origin.Lon, r.Origin.Lat),
		Destination:  logistics.NewGeoPoint(r.Destination.Lon, r.Destination.Lat),
		VehiclePlate: demoPlateFor(r.ID),
		ADRClass:     "",
		ATPClass:     "",
	}
}

func countryOf(dest string) string {
	switch dest {
	case "München Riem Hub":
		return "DE"
	default:
		return "IT"
	}
}

func demoPlateFor(id string) string {
	switch id {
	case VeronaMilano.ID:
		return "FA123LT"
	case VeronaNapoli.ID:
		return "FB456TR"
	case VeronaMunchen.ID:
		return "FC789BR"
	default:
		return "FZ000XX"
	}
}
