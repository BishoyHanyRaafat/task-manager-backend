package dto

import "task_manager/public/repositories/models"

// These aliases exist to keep Swagger/OpenAPI schemas named and stable
// without duplicating wrapper structs everywhere.

// AuthTokenEnvelope is the unified envelope for endpoints returning JWT tokens.
// It avoids inline/anonymous schemas in the generated OpenAPI spec.
type AuthTokenEnvelope = Envelope[AuthTokenResponse]

// LogoutEnvelope is the unified envelope for logout responses.
// It avoids inline/anonymous schemas in the generated OpenAPI spec.
type LogoutEnvelope = Envelope[LogoutResponse]

// MeEnvelope is the unified envelope for "current user" responses.
// It avoids inline/anonymous schemas in the generated OpenAPI spec.
type MeEnvelope = Envelope[MeResponse]

type (
	TeamsEnvelope            = Envelope[[]models.Team]
	TeamsInvitationsEnvelope = Envelope[models.Invitation]
	TeamsTaskEnvelope        = Envelope[models.Task]
	UserTeamsEnvelope        = Envelope[models.UserTeam]
)
