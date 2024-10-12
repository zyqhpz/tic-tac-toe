package main

import (
	"html/template"
	"net/http"
	"os"
	reader "tic-tac-toe/pkg"
	logger "tic-tac-toe/utils"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type TemplateData struct {
	WebSocketUrl string
	Board        [][]string
	PlayerID     string
	MatchID      string
}

func main() {

	logger.InitializeLogger()

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	if env == "development" {
		err := godotenv.Load()
		if err != nil {
			logger.Log().Fatal().Err(err).Msg("Error loading .env file")
		}
	}

	websocketUrl := os.Getenv("WEBSOCKET_URL")

	// Initialize the Tic Tac Toe board
	initialBoard := reader.EmptyBoard

	// Handle static files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./views/js/"))))

	// Handle static files for images
	http.Handle("/icons/", http.StripPrefix("/icons/", http.FileServer(http.Dir("./views/icons/"))))

	temp_uuid := ""
	temp_matchID := ""

	// Serve the main HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create a TemplateData instance with the WebSocketHost and initial board
		temp_uuid = uuid.New().String()
		temp_matchID = reader.GenerateUniqueMatchID()
		data := TemplateData{
			WebSocketUrl: websocketUrl,
			Board:        initialBoard,
			PlayerID:     temp_uuid,
			MatchID:      temp_matchID,
		}

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
	http.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {
		reader.SocketReaderCreate(w, r, temp_uuid, temp_matchID)
	})

	var PORT = "8080"
	logger.Log().Info().Msg("Server started at port " + PORT)

	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		logger.Log().Fatal().Err(err).Msg("Error starting server")
	}
}
