package oauth

import (
	"context"
	"net/http"
	"strings"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"
)

// NewStartHandler creates a handler that starts the Google OAuth consent flow.
//
// Summary:
// - Validates OAuth flow availability.
// - Generates the Google consent URL and redirects the caller.
//
// Attributes:
// - flow: OAuth authorization flow used to initialize the consent request.
//
// Returns:
// - gin.HandlerFunc: Handler that redirects to Google or returns an API error.
//
// NewStartHandler godoc
// @Summary Start Google OAuth consent flow
// @Tags auth
// @Produce json
// @Success 302 {string} string "Redirect to Google consent screen"
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/auth/google/start [get]
func NewStartHandler(flow *AuthorizationFlow) gin.HandlerFunc {
	return func(c *gin.Context) {
		if flow == nil {
			api.WriteError(c, api.NewHTTPError(
				http.StatusServiceUnavailable,
				api.ErrorCodeGoogleServiceUnavailable,
				"OAuth Google n'est pas configure",
			))
			return
		}

		authURL, _, err := flow.Start()
		if err != nil {
			api.WriteError(c, api.NewHTTPError(
				http.StatusInternalServerError,
				api.ErrorCodeInternalServerError,
				"Echec generation URL consentement Google",
			))
			return
		}

		c.Redirect(http.StatusFound, authURL)
	}
}

// NewCallbackHandler creates a handler that processes the OAuth callback.
//
// Summary:
// - Delegates callback handling to `NewCallbackHandlerWithHook` without a post-connect hook.
//
// Attributes:
// - flow: OAuth authorization flow used to validate state and exchange code.
//
// Returns:
// - gin.HandlerFunc: Handler that finalizes OAuth consent and returns connection status.
//
// NewCallbackHandler godoc
// @Summary Handle Google OAuth callback
// @Tags auth
// @Produce json
// @Param state query string true "OAuth state"
// @Param code query string true "OAuth authorization code"
// @Success 200 {object} map[string]string
// @Failure 400 {object} api.ErrorEnvelope
// @Failure 503 {object} api.ErrorEnvelope
// @Router /api/v1/auth/google/callback [get]
func NewCallbackHandler(flow *AuthorizationFlow) gin.HandlerFunc {
	return NewCallbackHandlerWithHook(flow, nil)
}

// NewCallbackHandlerWithHook creates a callback handler with an optional success hook.
//
// Summary:
// - Validates required callback query parameters.
// - Completes the OAuth callback flow and optionally runs a post-connection hook.
//
// Attributes:
// - flow: OAuth authorization flow used to validate state and exchange code.
// - onConnected: Optional callback executed after a successful OAuth connection.
//
// Returns:
// - gin.HandlerFunc: Handler that returns API errors or `{ "status": "connected" }`.
func NewCallbackHandlerWithHook(flow *AuthorizationFlow, onConnected func(context.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if flow == nil {
			api.WriteError(c, api.NewHTTPError(
				http.StatusServiceUnavailable,
				api.ErrorCodeGoogleServiceUnavailable,
				"OAuth Google n'est pas configure",
			))
			return
		}

		state := strings.TrimSpace(c.Query("state"))
		if state == "" {
			api.WriteError(c, api.NewHTTPError(
				http.StatusBadRequest,
				api.ErrorCodeInvalidParameters,
				"Le parametre 'state' est requis",
			))
			return
		}

		code := strings.TrimSpace(c.Query("code"))
		if code == "" {
			api.WriteError(c, api.NewHTTPError(
				http.StatusBadRequest,
				api.ErrorCodeInvalidParameters,
				"Le parametre 'code' est requis",
			))
			return
		}

		if err := flow.Callback(c.Request.Context(), state, code); err != nil {
			api.WriteError(c, api.NewHTTPError(
				http.StatusBadRequest,
				api.ErrorCodeGoogleServiceUnavailable,
				"Echec consentement Google",
			))
			return
		}

		if onConnected != nil {
			onConnected(c.Request.Context())
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "connected",
		})
	}
}
