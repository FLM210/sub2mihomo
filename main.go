package main

import (
	"log"
	"net/http"

	"sub2mihomo/internal/handlers"
)

func main() {
	http.HandleFunc("/convert", handlers.ConvertHandler)
	http.HandleFunc("/", handlers.HomeHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
