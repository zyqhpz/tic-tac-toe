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
}

type Player struct {
	ID   string `json:"userId"`
	Mark string `json:"mark"`
}

type Move struct {
	Row     int    `json:"row"`
	Col     int    `json:"col"`
	Player  Player `json:"player"`
	MatchID string `json:"matchId"`
	Type    string `json:"type"`
}

var (
	savedsocketreader = make([]*socketReader, 0) // List of all socket connections
	mutex             = &sync.Mutex{}            // Mutex to handle concurrent access
)

func (i *socketReader) broadcast(move Move) {
	mutex.Lock()
	defer mutex.Unlock()

	// Loop over all socket readers and broadcast the move
	for _, g := range savedsocketreader {
		// Check if the socket is still open before writing
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
	log.Printf("New connection added. Total active connections: %d", len(savedsocketreader))
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

// Handler to create and manage new WebSocket connections
func SocketReaderCreate(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	reader := &socketReader{con: conn}
	reader.addReader()

	go reader.read()
}
