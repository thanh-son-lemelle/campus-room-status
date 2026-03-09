package main

import (
    "campus-room-status/internal/health"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    health.RegisterRoutes(r)
    r.Run(":8080")
}