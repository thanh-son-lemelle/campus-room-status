package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/swaggo/swag/cmd/swag@v1.8.12 init --generalInfo ../../cmd/api/main.go --dir ../../cmd/api,../../internal/api,../../internal/app,../../internal/google/oauth,../../internal/health,../../internal/buildings,../../internal/rooms --output . --outputTypes json --parseInternal
//go:embed swagger.json
var swaggerJSON []byte

func SpecJSON() []byte {
	out := make([]byte, len(swaggerJSON))
	copy(out, swaggerJSON)
	return out
}

func NewOpenAPIHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", swaggerJSON)
	}
}
