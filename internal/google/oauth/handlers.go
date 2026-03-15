package oauth

import (
	"net/http"
	"strings"

	"campus-room-status/internal/api"
	"github.com/gin-gonic/gin"
)

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

func NewCallbackHandler(flow *AuthorizationFlow) gin.HandlerFunc {
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
				"Echec consentement Google: "+err.Error(),
			))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "connected",
		})
	}
}
