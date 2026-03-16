package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/swaggo/swag/cmd/swag@v1.8.12 init --generalInfo ../../cmd/api/main.go --dir ../../cmd/api,../../internal/api,../../internal/app,../../internal/google/oauth,../../internal/health,../../internal/buildings,../../internal/rooms --output . --outputTypes json --parseInternal
//go:embed swagger.json
var swaggerJSON []byte

// SpecJSON specs json.
//
// Summary:
// - Specs json.
//
// Attributes:
// - None.
//
// Returns:
// - value1 ([]byte): Returned value.
func SpecJSON() []byte {
	out := make([]byte, len(swaggerJSON))
	copy(out, swaggerJSON)
	return out
}

// NewOpenAPIHandler creates a new open api handler.
//
// Summary:
// - Creates a new open api handler.
//
// Attributes:
// - None.
//
// Returns:
// - value1 (gin.HandlerFunc): Returned value.
func NewOpenAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", swaggerJSON)
	}
}
