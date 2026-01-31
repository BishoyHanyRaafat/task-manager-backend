package response

// SignupRequest is used by POST /auth/signup
type SignupRequest struct {
	FirstName string `json:"firstname" validate:"required,min=2,max=100"`
	LastName  string `json:"lastname" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

// LoginRequest is used by POST /auth/login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthTokenData struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

type MeData struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}
