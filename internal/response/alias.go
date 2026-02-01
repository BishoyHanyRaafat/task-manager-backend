package response

// These aliases exist to keep Swagger/OpenAPI schemas named and stable
// without duplicating wrapper structs everywhere.

// AuthTokenEnvelope is the unified envelope for endpoints returning JWT tokens.
// It avoids inline/anonymous schemas in the generated OpenAPI spec.
type AuthTokenEnvelope = Envelope[AuthTokenData]

// LogoutEnvelope is the unified envelope for logout responses.
// It avoids inline/anonymous schemas in the generated OpenAPI spec.
type LogoutEnvelope = Envelope[LogoutResponse]

// MeEnvelope is the unified envelope for "current user" responses.
// It avoids inline/anonymous schemas in the generated OpenAPI spec.
type MeEnvelope = Envelope[MeData]

