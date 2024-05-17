package inspector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rodweb/webhook-inspector/internal/config"
	"github.com/rodweb/webhook-inspector/internal/request"
	"log"
	"net/http"
)

// TODO: add mutex
var clients = make(map[*Client]bool)

type Client struct {
	events chan string
}

type Options struct {
	// TODO: port as int
	Port string
}

func Start(ctx context.Context, errChan chan<- error, reqChan <-chan request.Request, opts Options) func(context.Context) {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", opts.Port),
		Handler: routes(),
	}

	go func() {
		log.Printf("Inspector listening at http://localhost:%s\n", opts.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	go func() {
		for {
			select {
			case req := <-reqChan:
				broadcast(req, clients)
			case <-ctx.Done():
				return
			}
		}
	}()

	shutdownFunc := func(ctx context.Context) {
		_ = server.Shutdown(ctx)
	}

	return shutdownFunc
}

func routes() *http.ServeMux {
	mux := http.NewServeMux()
	appHandler := http.FileServer(http.Dir("public"))
	mux.Handle("/", appHandler)
	mux.HandleFunc("/sse", sseHandler)
	return mux
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	client := &Client{}
	clients[client] = true
	defer delete(clients, client)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-control")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client.events = make(chan string)
	defer close(client.events)

	ctx := r.Context()

	for {
		select {
		case event := <-client.events:
			_, err := fmt.Fprintf(w, "data: %s\n\n", event)
			if err != nil {
				log.Println("failed to write event")
				continue
			}
			w.(http.Flusher).Flush()
		case <-ctx.Done():
			return
		}
	}
}

func OptionsFromConfig(config config.Config) Options {
	opts := Options{
		Port: "8080",
	}
	if config.Port != "" {
		opts.Port = config.Port
	}
	return opts
}

func broadcast(req request.Request, clients map[*Client]bool) {
	if len(clients) == 0 {
		return
	}

	data, err := json.Marshal(req)
	if err != nil {
		log.Println("failed to marshal request")
		return
	}

	log.Printf("Broadcasting request to %d clients\n", len(clients))
	for client := range clients {
		client.events <- string(data)
	}
}
