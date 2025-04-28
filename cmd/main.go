package main

import (
	"go-learning/internal/api"
	"go-learning/internal/storage"
)

func main() {
	store, err := storage.NewMemoryStore("./persons.db", 10000)
	if store == nil || err != nil {
		panic("Failed to create memory store")
	}

	api_handler := api.NewAPI(store)
	api_handler.Start(8080) // Start the API server on port 8080
}
