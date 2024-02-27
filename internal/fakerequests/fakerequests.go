package fakerequests

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rodweb/webhook-inspector/internal/request"
	"time"
)

var fakeMethods = []string{"GET", "POST", "PUT", "DELETE"}
var fakeEndpoints = []string{"/users", "/products", "/orders", "/payments"}
var fakeNames = []string{"John Doe", "Jane Doe", "John Smith", "Jane Smith"}

func Start(ctx context.Context, reqChan chan<- request.Request) {
	go func() {
		// every 5 seconds
		for {
			// id is a random string
			id := uuid.New()

			method := fakeMethods[time.Now().Unix()%int64(len(fakeMethods))]
			endpoint := fakeEndpoints[time.Now().Unix()%int64(len(fakeEndpoints))]
			name := fakeNames[time.Now().Unix()%int64(len(fakeNames))]
			age := time.Now().Unix() % 100

			req := request.Request{
				ID:       id.String(),
				Method:   method,
				Endpoint: endpoint,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
					"User-Agent":   {"Go-http-client/1.1"},
				},
				Body:      fmt.Sprintf("{\"name\": \"%s\", \"age\": %d}", name, age),
				Timestamp: time.Now().UnixMilli(),
			}

			select {
			case reqChan <- req:
				time.Sleep(5 * time.Second)
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
}
