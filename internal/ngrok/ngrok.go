package ngrok

import (
	"context"
	"github.com/google/uuid"
	"github.com/rodweb/webhook-inspector/internal/config"
	"github.com/rodweb/webhook-inspector/internal/request"
	"golang.ngrok.com/ngrok"
	ngrokConfig "golang.ngrok.com/ngrok/config"
	"io"
	"log"
	"net/http"
	"time"
)

type Options struct {
	AuthToken string
	Domain    string
}

func Start(ctx context.Context, errChan chan<- error, reqChan chan<- request.Request, opts Options) func() {
	var endpointOptions ngrokConfig.HTTPEndpointOption
	if opts.Domain != "" {
		endpointOptions = ngrokConfig.WithDomain(opts.Domain)
	}

	var connectOption ngrok.ConnectOption
	if opts.AuthToken != "" {
		connectOption = ngrok.WithAuthtoken(opts.AuthToken)
	}

	tunnel, err := ngrok.Listen(ctx,
		ngrokConfig.HTTPEndpoint(endpointOptions),
		connectOption,
	)
	if err != nil {
		log.Printf("Failed to create tunnel: %s\n", err)
		errChan <- err
		return nil
	}

	go func() {
		log.Printf("Tunnel listening at %s\n", tunnel.URL())
		err := http.Serve(tunnel, NewHandler(reqChan))
		if err != nil {
			errChan <- err
		}
	}()

	shutdownFunc := func() {
		_ = tunnel.Close()
	}

	return shutdownFunc
}

type Handler struct {
	reqChan chan<- request.Request
}

func NewHandler(reqChan chan<- request.Request) *Handler {
	return &Handler{reqChan}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received")

	var body string
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %s\n", err)
		body = "Failed to read request body"
	} else if len(data) > 0 {
		body = string(data)
	}

	req := request.Request{
		ID:        uuid.New().String(),
		Method:    r.Method,
		Endpoint:  r.URL.String(),
		Headers:   r.Header,
		Body:      body,
		Timestamp: time.Now().UnixMilli(),
	}

	h.reqChan <- req

	w.WriteHeader(http.StatusOK)
	log.Println("Response sent")
}

func OptionsFromConfig(config config.Config) Options {
	opts := Options{}
	if config.Token != "" {
		opts.AuthToken = config.Token
	}
	if config.Domain != "" {
		opts.Domain = config.Domain
	}
	return opts
}
