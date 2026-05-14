package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/open-workout/ow/services/user-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/user-service/internal/service"
	"github.com/open-workout/ow/services/user-service/internal/transport/http/handlers"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_DSN"))
	if err != nil {
		log.Fatalf("failed to open postgres connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}

	repo := repository.NewSqlRepository(db)
	svc := service.NewService(repo)
	userHandler := handlers.NewUserHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /users/{id}", userHandler.GetUser)
	mux.HandleFunc("PUT /users/{id}", userHandler.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", userHandler.DeleteUser)
	mux.HandleFunc("PUT /users/{id}/split", userHandler.UpdateSplit)

	port := 8081
	if p := os.Getenv("USER_SERVICE_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("user-service listening on :%d", port)
	log.Fatal(srv.ListenAndServe())
}
