package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/open-workout/ow/services/user-service/internal/transport/http/handlers"
)

func main() {
	port := 8081
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	userHandler := handlers.NewUserHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", userHandler.GetUser)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("user-service listening on :%d", port)
	log.Fatal(srv.ListenAndServe())
}
