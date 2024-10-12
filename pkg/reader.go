package pkg

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type socketReader struct {
	con *websocket.Conn
	id  string
}

type Game struct {
	MatchID string     `json:"matchId"`
	Board   [][]string `json:"board"`
}

type Player struct {
	ID   string `json:"userId"`
	Mark string `json:"mark"`
}

type Move struct {
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	Player Player `json:"player"`
	Game   Game   `json:"game"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

var (
	savedsocketreader = make([]*socketReader, 0) // List of all socket connections
	mutex             = &sync.Mutex{}            // Mutex to handle concurrent access
)

var mark = []string{"X", "O"}

func (i *socketReader) broadcast(move Move) {
	mutex.Lock()
	defer mutex.Unlock()

	// Loop over all socket readers and broadcast the move
	log.Printf("Broadcasting move: %+v", move)
	for i, g := range savedsocketreader {
		// Check if the socket is still open before writing
		if move.Type == "start" {
			move.Player.Mark = mark[i]
		}
		err := g.writeMsg(move)
		if err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			// Optionally remove closed/disconnected socket from the list
			continue
		}
	}
}

func (i *socketReader) read() {
	for {
		_, b, err := i.con.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var move Move
		if err := json.Unmarshal(b, &move); err == nil {
			log.Printf("Received move: %+v", move)

			// Broadcast the move to all clients, including the current player
			i.broadcast(move)

			winner := i.analyzeBoard(move.Game.Board)
			log.Println("Winner: ", winner)
			if winner != "" {
				move.Type = "end"
				move.Status = winner

				// Broadcast the end game message to all clients]
				i.broadcast(move)
			}
		} else {
			log.Printf("Error unmarshalling move: %v", err)
		}
	}

	// Remove the socket reader if the connection closes
	i.removeReader()
}

func (i *socketReader) writeMsg(move Move) error {
	msg, err := json.Marshal(move)
	if err != nil {
		return err
	}

	err = i.con.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Printf("Error writing message: %v", err)
		i.con.Close() // Close the connection if write fails
		return err
	}

	return nil
}

// Adds the current socket reader to the list of active connections
func (i *socketReader) addReader() {
	mutex.Lock()
	defer mutex.Unlock()
	savedsocketreader = append(savedsocketreader, i)
	log.Printf("New connection added with ID: %s. Total active connections: %d", i.id, len(savedsocketreader))

	if len(savedsocketreader) == 2 {
		// Send a message to the clients to start the game
		move := Move{
			Row: 0,
			Col: 0,
			Player: Player{
				ID:   "server",
				Mark: "X",
			},
			Type: "start",
		}
		go i.broadcast(move)
	}

}

// Removes the socket reader when a connection is closed
func (i *socketReader) removeReader() {
	mutex.Lock()
	defer mutex.Unlock()

	for idx, g := range savedsocketreader {
		if g == i {
			// Remove the socket reader from the list
			savedsocketreader = append(savedsocketreader[:idx], savedsocketreader[idx+1:]...)
			log.Printf("Connection closed. Total active connections: %d", len(savedsocketreader))
			break
		}
	}
}

// Analyze the board to check for a winner or a draw
func (i *socketReader) analyzeBoard(board [][]string) string {
	// Check rows
	for i := 0; i < 3; i++ {
		if board[i][0] == board[i][1] && board[i][1] == board[i][2] && board[i][0] != "" {
			return board[i][0]
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if board[0][i] == board[1][i] && board[1][i] == board[2][i] && board[0][i] != "" {
			return board[0][i]
		}
	}

	// Check diagonals
	if board[0][0] == board[1][1] && board[1][1] == board[2][2] && board[0][0] != "" {
		return board[0][0]
	}

	if board[0][2] == board[1][1] && board[1][1] == board[2][0] && board[0][2] != "" {
		return board[0][2]
	}

	// Check for a draw
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "" {
				return ""
			}
		}
	}

	return "draw"
}

// Handler to create and manage new WebSocket connections
func SocketReaderCreate(w http.ResponseWriter, r *http.Request, uuid string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	reader := &socketReader{con: conn, id: uuid}
	reader.addReader()

	go reader.read()
}
