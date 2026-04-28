package rentri

// Environment-level endpoint constants for the RENTRI service.
//
// The sandbox environment is published at demoapi.rentri.gov.it for
// the API surface and demobackoffice.rentri.gov.it for the human
// portal. The production environment is at api.rentri.gov.it. These
// hosts are stated in the official portal at https://www.rentri.gov.it
// and in the OpenAPI documentation at
// https://api.rentri.gov.it/docs and https://demoapi.rentri.gov.it/docs.
//
// Access dates: 2026-04-28.
const (
	SandboxAPIBase    = "https://demoapi.rentri.gov.it"
	SandboxPortalBase = "https://demobackoffice.rentri.gov.it"
	ProductionAPIBase = "https://api.rentri.gov.it"

	// OpenAPISpecPath is the documented OpenAPI v1.0 surface for the
	// "dati registri" namespace. The same path resolves under both
	// sandbox and production hosts.
	OpenAPISpecPath = "/docs/dati-registri/v1.0"
)

// Environment selects which base URL the Client adapter uses.
type Environment string

const (
	EnvSandbox    Environment = "sandbox"
	EnvProduction Environment = "production"
)

// BaseURL returns the API base for the supplied environment, or the
// sandbox base for any unrecognised value. The platform's default is
// always sandbox until the trasportatore's RENTRI certificate has
// been validated against a successful demoapi call.
func (e Environment) BaseURL() string {
	if e == EnvProduction {
		return ProductionAPIBase
	}
	return SandboxAPIBase
}

// PortalURL returns the human-portal base URL for the environment.
// The production portal uses the same host as the API surface.
func (e Environment) PortalURL() string {
	if e == EnvProduction {
		return ProductionAPIBase
	}
	return SandboxPortalBase
}
