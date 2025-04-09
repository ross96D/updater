package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/ross96D/updater/share/utils"
)

type webSocketHandler struct {
	upgrader websocket.Upgrader
}

func (wsh webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "file")
	if filename == "" {
		w.WriteHeader(400)
		w.Write([]byte("filename is emtpy"))
		return
	}

	path := filepath.Join(utils.TempDirectory(), filename)
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(404)
		} else {
			http.Error(w, err.Error(), 500)
		}
		return
	}
	defer file.Close()

	conn, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = writeData(conn, file)
	if err != nil {
		return
	}

	chn, err := watchFile(r.Context(), path, file)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(15*time.Second))
			if err != nil {
				return
			}
		case resp := <-chn:
			if resp.err != nil {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
			err = conn.WriteMessage(websocket.TextMessage, resp.data)
			if err != nil {
				return
			}
		}
	}
}

type message struct {
	err  error
	data []byte
}

func watchFile(ctx context.Context, path string, file *os.File) (<-chan message, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	err = w.Add(filepath.Join(utils.TempDirectory()))
	if err != nil {
		w.Close()
		return nil, err
	}

	chn := make(chan message)
	go func(chn chan<- message) {
		defer w.Close()
		for {
			select {
			case event := <-w.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					println("-------------------------- RECIEVE fsnotify.Write  -------------------------------")
					// read file and look for diff with repect of previous file
					// file.Seek(offset, 0)
					data, err := io.ReadAll(file)
					print("------------------ DATA LEN ", len(data), "----------------")
					chn <- message{err: err, data: data}
				}
			case <-ctx.Done():
				return
			}
		}
	}(chn)
	return chn, nil
}

func writeData(conn *websocket.Conn, file *os.File) error {
	var buffer [4086]byte
	for {
		n, err := file.Read(buffer[:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				println("-------------------------- END writeData -------------------------------")
				return nil
			} else {
				return err
			}
		}
		println("-------------------------- SEND MESSAGE IN WRITE DATA  -------------------------------")
		conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
		err = conn.WriteMessage(websocket.TextMessage, buffer[0:n])
		if err != nil {
			return err
		}
	}
}
