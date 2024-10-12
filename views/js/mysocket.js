class MySocket {
  constructor() {
    this.mysocket = null;
    this.userId = "";
    this.playerMark = ""; // This will store the player's mark (e.g., 'X' or 'O')
    this.gameId = ""; // Store the current Game ID
    this.matchId = ""; // Store the current Match ID
  }

  connectSocket() {
    console.log("connecting socket");
    const host = WEBSOCKET_URL;
    var socket = new WebSocket(host + "/socket");
    this.mysocket = socket;

    socket.onmessage = (e) => {
      const data = JSON.parse(e.data);
      this.handleMessage(data);
    };
    socket.onopen = () => {
      console.log("socket opened");
    };
    socket.onclose = () => {
      console.log("socket closed");
    };
  }

  sendMove(row, col) {
    // Create a move object that matches the Go structure
    const move = {
      row: row,
      col: col,
      player: {
        userId: this.userId, // The player's unique ID
        mark: this.playerMark, // The player's mark, e.g., 'X' or 'O'
      },
      userId: this.userId, // Redundant, but included as in Go structure
      matchId: this.matchId, // The current match ID
      type: "move", // Specify the type of message
    };

    // Send the move object as a JSON string
    this.mysocket.send(JSON.stringify(move));
  }

  handleMessage(data) {
    console.log("message received", data);
    if (data.type === "move") {
      // Update the board based on the message
      this.updateBoard(data.row, data.col, data.player);
    }
    if (data.type === "status") {
      // Update the game status message
      document.getElementById("game-status").innerText = data.message;
    }
  }

  updateBoard(row, col, player) {
    const cell = document.querySelector(
      `.cell[data-row="${row}"][data-col="${col}"]`
    );
    if (cell) {
      cell.innerHTML = player; // Update the cell with the player's symbol

      // Optionally, disable further clicks on the cell
      cell.onclick = null; // Prevent further clicks
      cell.classList.add("bg-gray-300"); // Change background color to indicate it's occupied

      // Update game status (if applicable)
      this.updateGameStatus(player); // Call a function to update the game status
    }
  }

  // Function to update the game status message
  updateGameStatus(player) {
    const statusElement = document.getElementById("game-status");
    if (statusElement) {
      statusElement.innerText = `${player} has made a move!`; // Customize this message as needed
    }
  }
}

// Initialize WebSocket connection
const socket = new MySocket();
socket.connectSocket();
