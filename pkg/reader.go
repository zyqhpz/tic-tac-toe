package pkg

import (
	"encoding/json"
	"net/http"
	"sync"
	logger "tic-tac-toe/utils"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var EmptyBoard = [][]string{
	{"", "", ""},
	{"", "", ""},
	{"", "", ""},
}

type socketReader struct {
	con    *websocket.Conn
	id     string
	player Player
}

type Game struct {
	MatchID string     `json:"matchId"`
	Board   [][]string `json:"board"`
	Players []Player   `json:"players"`
	Status  string     `json:"status"`
}

type Player struct {
	ID             string `json:"userId"`
	InitialMatchID string `json:"initialMatchId"`
	Mark           string `json:"mark"`
}

type Message struct {
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	Player Player `json:"player"`
	Game   Game   `json:"game"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

var (
	savedsocketreader = make([]*socketReader, 0) // List of all socket connections
	mutex             = &sync.Mutex{}
	Games             = make(map[string]Game)
)

var mark = []string{"X", "O"}

func (i *socketReader) broadcast(message Message) {
	mutex.Lock()
	defer mutex.Unlock()

	// Loop over all socket readers and broadcast the
	logger.Log().Debug().Msgf("Broadcasting message: %+v", message)
	for i, g := range savedsocketreader {
		if message.Type == "start" {
			message.Player.Mark = mark[i]
		}
		err := g.writeMsg(message)
		if err != nil {
			logger.Log().Error().Err(err).Msg("Error broadcasting to client")
			continue
		}
	}
}

func (i *socketReader) startGameMessage(matchID string) {
	mutex.Lock()
	defer mutex.Unlock()

	message := Message{
		Player: Player{
			Mark: "",
		},
		Type: "start",
	}

	player_counter := 0

	// Loop over all socket readers and only send the start message to the players in the game
	for _, g := range savedsocketreader {
		logger.Log().Debug().Msgf("Player ID: %s, InitialMatchID: %s", g.id, g.player.InitialMatchID)
		if g.player.InitialMatchID == matchID {
			message.Player.Mark = mark[player_counter]
			player_counter++
			err := g.writeMsg(message)
			logger.Log().Debug().Msgf("Sending start message to player: %s", g.id)
			if err != nil {
				logger.Log().Error().Err(err).Msg("Error starting game")
				continue
			}

			if player_counter == 2 {
				break
			}
		}
	}

	logger.Log().Info().Msgf("Game started for matchID: %s", matchID)
}

func endGameMessage(matchID string) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := Games[matchID]; !exists {
		logger.Log().Error().Msg("MatchID does not exist")
		return
	}

	game := Games[matchID]
	if game.Status != "ended" {
		message := Message{
			Type:   "end",
			Status: "closed",
			Game: Game{
				MatchID: matchID,
			},
		}

		logger.Log().Info().Msgf("Ending game for matchID: %s", matchID)

		// Loop over all socket readers and only send the end message to the players in the game
		for _, g := range savedsocketreader {
			if g.player.InitialMatchID == matchID {
				err := g.writeMsg(message)
				if err != nil {
					logger.Log().Error().Err(err).Msg("Error ending game")
					continue
				}
			}
		}

		delete(Games, matchID)
		logger.Log().Info().Msgf("Game ended for matchID: %s", matchID)
	}
}

func (i *socketReader) read() {
	for {
		temp_matchID := i.player.InitialMatchID
		_, b, err := i.con.ReadMessage()
		if err != nil {
			logger.Log().Error().Err(err).Msg("Error reading message")
			endGameMessage(temp_matchID)
			break
		}

		var message Message
		if err := json.Unmarshal(b, &message); err == nil {
			logger.Log().Debug().Msgf("Received message: %+v", message)

			if message.Type == "join" {
				// check if matchID exists
				if _, exists := Games[message.Game.MatchID]; exists {
					// Add the player to the game
					game := Games[message.Game.MatchID]

					// Check if the game is already full
					if len(game.Players) == 2 {
						logger.Log().Info().Msgf("Game %s is already full", message.Game.MatchID)
						message.Type = "join"
						message.Status = "full"
						i.writeMsg(message)
					} else {
						delete(Games, i.player.InitialMatchID)

						i.player = message.Player
						game.Players = append(game.Players, i.player)
						Games[message.Game.MatchID] = game

						logger.Log().Debug().Msgf("Player joined the game. Total players: %d", len(Games[message.Game.MatchID].Players))

						// Send a message to the clients to start the game
						i.startGameMessage(message.Game.MatchID)
					}
				} else {
					logger.Log().Error().Msg("MatchID does not exist")
					message.Type = "join"
					message.Status = "failed"
					i.writeMsg(message)
				}
			}

			if message.Type == "move" {
				// Broadcast the message to all clients, including the current player
				i.broadcast(message)
				winner := i.analyzeBoard(message.Game.Board)
				if winner != "" {
					logger.Log().Info().Msgf("Game over. Winner: %s", winner)
					message.Type = "end"
					message.Status = winner

					game := Games[message.Game.MatchID]
					game.Status = "ended"
					Games[message.Game.MatchID] = game

					// Broadcast the end game message to all clients
					i.broadcast(message)
				}
			}
		} else {
			logger.Log().Error().Err(err).Msg("Error unmarshalling message")
		}
	}

	// Remove the socket reader if the connection closes
	i.removeReader()
}

func (i *socketReader) writeMsg(message Message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = i.con.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		logger.Log().Error().Err(err).Msg("Error writing message")
		i.con.Close()
		return err
	}

	return nil
}

// Adds the current socket reader to the list of active connections
func (i *socketReader) addReader() {
	mutex.Lock()
	defer mutex.Unlock()

	i.player.ID = i.id

	// Create a new game
	game := Game{
		MatchID: i.player.InitialMatchID,
		Board:   EmptyBoard,
		Players: []Player{
			i.player,
		},
		Status: "waiting",
	}

	// Add the game to the list of active games
	Games[game.MatchID] = game

	savedsocketreader = append(savedsocketreader, i)
	logger.Log().Info().Msgf("New connection added with ID: %s. Total active connections: %d", i.id, len(savedsocketreader))
	logger.Log().Info().Msgf("New game created with matchID: %s", game.MatchID)
}

// Removes the socket reader when a connection is closed
func (i *socketReader) removeReader() {
	mutex.Lock()
	defer mutex.Unlock()

	for idx, g := range savedsocketreader {
		if g == i {
			// Remove the socket reader from the list
			savedsocketreader = append(savedsocketreader[:idx], savedsocketreader[idx+1:]...)
			logger.Log().Info().Msgf("Connection closed. Total active connections: %d", len(savedsocketreader))
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

func GenerateUniqueMatchID() string {
	var matchID string

	// Loop until a unique matchID is generated
	for {
		matchID = gofakeit.Animal()
		// Check if the matchID already exists in the games map
		if _, exists := Games[matchID]; !exists {
			break // if it's unique, exit the loop
		}
	}

	return matchID
}

// Handler to create and manage new WebSocket connections
func SocketReaderCreate(w http.ResponseWriter, r *http.Request, uuid string, matchID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log().Error().Err(err).Msg("Failed to upgrade connection")
		return
	}

	reader := &socketReader{con: conn, id: uuid, player: Player{InitialMatchID: matchID}}
	reader.addReader()

	go reader.read()
}
