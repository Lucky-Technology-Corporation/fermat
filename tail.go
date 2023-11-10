package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func tailLogsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade websocket connection", err.Error())
		http.Error(w, "Failed to upgrade to websocket connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	queryParams := r.URL.Query()
	tailFile := queryParams.Get("path")
	if !strings.HasSuffix(tailFile, ".log") {
		http.Error(w, "Can't tail non-log files", http.StatusBadRequest)
		return
	}

	t, err := tail.TailFile(tailFile, tail.Config{Follow: true, MustExist: true})
	if err != nil {
		log.Println("Failed to tail logs", err.Error())
		http.Error(w, "Failed to tail logs", http.StatusInternalServerError)
		return
	}
	defer t.Stop()

	clientClosed := r.Context().Done()

	for {
		select {
		case line := <-t.Lines:
			err := conn.WriteMessage(websocket.TextMessage, []byte(line.Text))
			if err != nil {
				log.Println("Error writing to websocket connection", err)
				return
			}
		case <-clientClosed:
			// Client closed connection
			return
		}
	}
}
