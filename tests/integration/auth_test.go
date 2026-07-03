package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

func TestAuth_RegisterLoginFlow(t *testing.T) {
	// This test exercises the core user flow:
	// Register -> Login -> Access protected endpoint (/api/v1/auth/me)

	router := newAuthRouter()

	// --- Register a new user ---
	tokens := registerUser(t, router, "flow@test.com", "FlowPass1")
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, "flow@test.com", tokens.UserEmail)
	assert.Equal(t, "customer", string(tokens.UserRole))

	// --- Login with the same credentials ---
	loginTokens := loginUser(t, router, "flow@test.com", "FlowPass1")
	assert.NotEmpty(t, loginTokens.AccessToken)
	assert.NotEmpty(t, loginTokens.RefreshToken)
	assert.Equal(t, tokens.UserID, loginTokens.UserID)

	// --- Access the protected /me endpoint with the login access token ---
	w := httpDo(t, router, http.MethodGet, "/api/v1/auth/me", "", loginTokens.AccessToken)
	require.Equal(t, http.StatusOK, w.Code)

	var meResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID    uint   `json:"id"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &meResp)
	require.NoError(t, err)
	assert.Equal(t, 0, meResp.Code)
	assert.Equal(t, "flow@test.com", meResp.Data.Email)
	assert.Equal(t, "customer", meResp.Data.Role)
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	router := newAuthRouter()

	// Register first time — should succeed.
	_ = registerUser(t, router, "dupe@test.com", "DupePass1")

	// Register with the same email — should fail with 409.
	w := httpDo(t, router, http.MethodPost, "/api/v1/auth/register",
		`{"email":"dupe@test.com","password":"DupePass1"}`, "")
	require.Equal(t, http.StatusConflict, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeEmailAlreadyExists, resp.Code)
	assert.Contains(t, resp.Message, "already registered")
}

func TestAuth_LoginWrongPassword(t *testing.T) {
	router := newAuthRouter()

	// Register a user first.
	_ = registerUser(t, router, "wrongpw@test.com", "RightPw1")

	// Login with wrong password — should fail with 401.
	w := httpDo(t, router, http.MethodPost, "/api/v1/auth/login",
		`{"email":"wrongpw@test.com","password":"WrongPw1"}`, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeInvalidCredentials, resp.Code)
	assert.Contains(t, resp.Message, "invalid email or password")
}

func TestAuth_ProtectedEndpointNoToken(t *testing.T) {
	router := newAuthRouter()

	// Access /me without any Authorization header.
	w := httpDo(t, router, http.MethodGet, "/api/v1/auth/me", "", "")
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeUnauthorized, resp.Code)
}

func TestAuth_ProtectedEndpointInvalidToken(t *testing.T) {
	router := newAuthRouter()

	// Access /me with a malformed token.
	w := httpDo(t, router, http.MethodGet, "/api/v1/auth/me", "", "this.is.not.a.valid.jwt")
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeUnauthorized, resp.Code)

	// Access /me with a token that uses a different secret.
	w2 := httpDo(t, router, http.MethodGet, "/api/v1/auth/me", "", "Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwidXNlcl9pZCI6MX0.invalid")
	require.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestAuth_RefreshTokenFlow(t *testing.T) {
	router := newAuthRouter()

	// 1. Register a new user.
	tokens := registerUser(t, router, "refresh@test.com", "RefPass1")

	// 2. Call /auth/refresh with the access token (for auth) and the refresh
	//    token (in the body) to get a new token pair.
	w := httpDo(t, router, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token":"`+tokens.RefreshToken+`"}`, tokens.AccessToken)
	require.Equal(t, http.StatusOK, w.Code)

	refreshed := parseAuthResponse(t, w.Body.Bytes())
	assert.NotEmpty(t, refreshed.AccessToken)
	assert.NotEmpty(t, refreshed.RefreshToken)
	assert.Equal(t, tokens.UserID, refreshed.UserID)

	// The new refresh token should be different from the original one (UUID-based rotation).
	assert.NotEqual(t, tokens.RefreshToken, refreshed.RefreshToken)

	// 3. Use the new access token to access a protected endpoint.
	w2 := httpDo(t, router, http.MethodGet, "/api/v1/auth/me", "", refreshed.AccessToken)
	require.Equal(t, http.StatusOK, w2.Code)

	// 4. The old refresh token should be invalid (rotated) so refreshing again
	//    with it should fail.
	w3 := httpDo(t, router, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token":"`+tokens.RefreshToken+`"}`, refreshed.AccessToken)
	require.Equal(t, http.StatusUnauthorized, w3.Code)

	var errResp response.Response
	err := json.Unmarshal(w3.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeUnauthorized, errResp.Code)
}
