package main

import (
	"log"
	"net/http"
	reader "webrtc-tic-tac-toe/pkg"
)

func main() {

	myhttp := http.NewServeMux()
	fs := http.FileServer(http.Dir("./views/"))
	myhttp.Handle("/", http.StripPrefix("", fs))

	myhttp.HandleFunc("/socket", reader.SocketReaderCreate)

	log.Println("http://localhost:8080")
	http.ListenAndServe(":8080", myhttp)
}
