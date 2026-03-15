package main

import (
	"log"

	"campus-room-status/internal/app"
)

// @title Campus Room Status API
// @version 1.0
// @description API de disponibilite des salles campus (buildings, rooms, schedules, health).
// @BasePath /
// @schemes http https
func main() {
	r := app.NewRouter()

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
