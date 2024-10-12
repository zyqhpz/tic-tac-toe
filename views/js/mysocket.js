class MySocket {
  constructor() {
    this.socket = null;
    this.userId = "";
    this.playerMark = ""; // This will store the player's mark (e.g., 'X' or 'O')
    this.gameId = ""; // Store the current Game ID
    this.matchId = ""; // Store the current Match ID
    this.isYourTurn = false; // Flag to check if it's the player's turn
    this.board = [
      ["", "", ""],
      ["", "", ""],
      ["", "", ""],
    ];
  }

  connectSocket() {
    if (this.socket) {
      console.warn("WebSocket is already connected.");
      return;
    }
    const host = WEBSOCKET_URL;
    this.socket = new WebSocket(host + "/socket");

    this.socket.onmessage = (e) => {
      const data = JSON.parse(e.data);
      this.handleMessage(data);
    };
    this.socket.onopen = () => {
      console.log("socket opened");
    };
    this.socket.onclose = () => {
      console.log("socket closed");
    };
  }

  sendMove(row, col) {
    if (this.isYourTurn === false) {
      alert("It's not your turn!");
      return;
    }
    // Create a move object that matches the Go structure
    row = Number(row);
    col = Number(col);
    const move = {
      row: row,
      col: col,
      player: {
        userId: this.userId, // The player's unique ID
        mark: this.playerMark, // The player's mark, e.g., 'X' or 'O'
      },
      game: {
        matchId: this.matchId, // The current match ID
        board: this.board, // The current game board
      },
      type: "move", // Specify the type of message
      status: "playing", // The game status
    };

    this.updateBoard(move.type, row, col, move.player);

    // Send the move object as a JSON string
    this.socket.send(JSON.stringify(move));
  }

  handleMessage(data) {
    if (data.type === "move") {
      if (data.player.userId != this.userId) {
        this.updateBoard(data.type, data.row, data.col, data.player);
      }
      this.updateGameStatus(data.player);
    }
    if (data.type === "start") {
      this.playerMark = data.player.mark;
      document.getElementById("player-mark").innerText =
        "You are player " + this.playerMark;

      const baseText = "Game started! ";

      if (data.player.mark === "X") {
        document.getElementById("game-status").innerText =
          baseText + "Your turn!";
        this.isYourTurn = true;
      } else {
        document.getElementById("game-status").innerText =
          baseText + "Waiting for opponent...";
      }

      // Update the board
      this.updateBoard(data.type, data.row, data.col, data.player);
    }
    if (data.type === "end") {
      if (data.status === "draw") {
        document.getElementById("game-status").innerText =
          "Game ended in a draw!";
      } else {
        if (data.player.userId === this.userId) {
          document.getElementById("game-status").innerText = "You won!";
        } else {
          document.getElementById("game-status").innerText = "You lost!";
        }
      }

      this.gameEnded();
    }
  }

  updateBoard(type, row, col, player) {
    if (type === "start") {
      // Clear the board
      this.restartGame();

      if (player.mark === "X") {
        this.isYourTurn = true;
      }

      if (this.isYourTurn) {
        // enable all cells that has empty mark
        const cells = document.querySelectorAll(".cell");
        cells.forEach((cell) => {
          if (cell.innerHTML === "") {
            cell.classList.remove("bg-red-100");
          }
        });
      }

      if (!this.isYourTurn) {
        // disable all cells that has empty mark
        const cells = document.querySelectorAll(".cell");
        cells.forEach((cell) => {
          if (cell.innerHTML === "") {
            cell.classList.add("bg-red-100");
          }
        });
      }
    }

    if (type === "move") {
      this.updateCell(row, col, player);
      if (player.userId === this.userId) {
        // disable all cells that has empty mark
        const cells = document.querySelectorAll(".cell");
        cells.forEach((cell) => {
          if (cell.innerHTML === "") {
            cell.classList.add("bg-red-100");
          }
        });
      } else {
        // enable all cells that has empty mark
        const cells = document.querySelectorAll(".cell");
        cells.forEach((cell) => {
          if (cell.innerHTML === "") {
            cell.classList.remove("bg-red-100");
          }
        });
      }
      this.isYourTurn = !this.isYourTurn;
    }
  }

  updateCell(row, col, player) {
    const cell = document.querySelector(
      `.cell[data-row="${row}"][data-col="${col}"]`
    );
    cell.innerHTML = player.mark;
    cell.onclick = null;
    if (player.mark === "X") {
      cell.classList.add("bg-blue-500");
    }
    if (player.mark === "O") {
      cell.classList.add("bg-green-500");
    }

    // update board
    this.board[row][col] = player.mark;
  }

  updateBoardStatus() {
    // update board to this.board
    const cells = document.querySelectorAll(".cell");
    cells.forEach((cell) => {
      const row = cell.getAttribute("data-row");
      const col = cell.getAttribute("data-col");
      cell.innerHTML = this.board[row][col];
      cell.onclick = () => {
        const row = cell.getAttribute("data-row");
        const col = cell.getAttribute("data-col");
        this.sendMove(row, col);
      };
    });
  }

  restartGame() {
    const cells = document.querySelectorAll(".cell");
    cells.forEach((cell) => {
      cell.innerHTML = "";
      cell.onclick = () => {
        const row = cell.getAttribute("data-row");
        const col = cell.getAttribute("data-col");
        this.sendMove(row, col);
      };
      cell.classList.remove("bg-gray-300");
      cell.classList.remove("bg-red-100");
      cell.classList.remove("bg-blue-500");
      cell.classList.remove("bg-green-500");
    });
  }

  gameEnded() {
    const cells = document.querySelectorAll(".cell");
    cells.forEach((cell) => {
      cell.onclick = null;
    });
  }

  // Function to update the game status message
  updateGameStatus(player) {
    const statusElement = document.getElementById("game-status");
    if (statusElement) {
      statusElement.innerText = `${player.mark} has made a move!`; // Customize this message as needed
    }
  }
}
