package auth_test

import (
	"net/http"
	"task_manager/internal/response"
	"task_manager/internal/testutil"
	"task_manager/internal/testutil/fakes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignupLoginMe_FakeRepo(t *testing.T) {
	repo := fakes.NewUserRepo()
	r := testutil.NewTestRouter(t, repo, "test-secret")

	// signup
	signupResp := testutil.DoJSON(t, r, http.MethodPost, "/api/v1/auth/signup", response.SignupRequest{
		FirstName: "Bishoy",
		LastName:  "Raafat",
		Email:     "bishoy@example.com",
		Password:  "password123",
	}, nil)
	require.Equal(t, http.StatusOK, signupResp.Code)

	signupEnv := testutil.DecodeJSON[response.EnvelopeAny](t, signupResp)
	require.True(t, signupEnv.Success)

	// login
	loginResp := testutil.DoJSON(t, r, http.MethodPost, "/api/v1/auth/login", response.LoginRequest{
		Email:    "bishoy@example.com",
		Password: "password123",
	}, nil)
	require.Equal(t, http.StatusOK, loginResp.Code)
	loginEnv := testutil.DecodeJSON[response.EnvelopeAny](t, loginResp)
	require.True(t, loginEnv.Success)

	dataMap, ok := loginEnv.Data.(map[string]any)
	require.True(t, ok)
	accessToken, _ := dataMap["access_token"].(string)
	require.NotEmpty(t, accessToken)

	// me (authorized)
	meResp := testutil.DoJSON(t, r, http.MethodGet, "/api/v1/me", nil, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	require.Equal(t, http.StatusOK, meResp.Code)
	meEnv := testutil.DecodeJSON[response.EnvelopeAny](t, meResp)
	require.True(t, meEnv.Success)
}

func TestMe_Unauthorized(t *testing.T) {
	repo := fakes.NewUserRepo()
	r := testutil.NewTestRouter(t, repo, "test-secret")

	meResp := testutil.DoJSON(t, r, http.MethodGet, "/api/v1/me", nil, nil)
	require.Equal(t, http.StatusUnauthorized, meResp.Code)

	meEnv := testutil.DecodeJSON[response.EnvelopeAny](t, meResp)
	require.False(t, meEnv.Success)
}
