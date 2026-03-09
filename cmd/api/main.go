package main

import (
    "log"
	"campus-room-status/internal/app"
)

func main() {
    r := app.NewRouter()

    if err := r.Run(":8080"); err != nil {
        log.Fatalf("failed to run server: %v", err)
    }
}