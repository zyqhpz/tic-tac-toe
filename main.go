package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	reader "tic-tac-toe/pkg"

	"github.com/joho/godotenv" // Import the package to read .env files
)

// Define a struct to hold the template data
type TemplateData struct {
	WebSocketUrl string
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the WebSocket host from environment variables
	websocketUrl := os.Getenv("WEBSOCKET_URL")

	// Handle static files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./views/js/"))))

	// Serve the main HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create a TemplateData instance with the WebSocketHost
		data := TemplateData{WebSocketUrl: websocketUrl}

		// Parse and execute the HTML template
		tmpl, err := template.ParseFiles("views/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Serve the WebSocket endpoint
	http.HandleFunc("/socket", reader.SocketReaderCreate)

	var PORT = "8080"
	log.Println("Server started at port " + PORT)

	// Start the server
	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		log.Fatal(err)
	}
}
