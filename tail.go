package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func tailLogsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	queryParams := r.URL.Query()
	tailFile := queryParams.Get("path")
	if !strings.HasSuffix(tailFile, ".log") {
		http.Error(w, "Can't tail non-log files", http.StatusBadRequest)
		return
	}

	numLines := 10
	initalLines := queryParams.Get("initial_lines")
	if initalLines != "" {
		numLines, err = strconv.Atoi(initalLines)
		if err != nil || numLines < 0 {
			http.Error(w, "Param inital_lines must be a non-negative integer", http.StatusBadRequest)
			return
		}
	}

	tailFile = filepath.Join("code", tailFile)

	// 1. Try and read intial_lines number lines from the file and send that in the first websocket message.
	cmd := exec.Command("tail", "-n", strconv.Itoa(numLines), tailFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Println("Failed to execute tail command", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade websocket connection", err.Error())
		return
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, out.Bytes())
	if err != nil {
		log.Println("Error writing to websocket connection", err)
		return
	}

	// 2. Start tailing the file from the end.
	t, err := tail.TailFile(tailFile, tail.Config{
		Follow:    true,
		MustExist: true,
		Location: &tail.SeekInfo{
			Whence: os.SEEK_END,
			Offset: 0,
		},
		Logger: tail.DiscardingLogger,
	})
	if err != nil {
		log.Println("Failed to tail logs", err.Error())
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
