package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/rodweb/webhook-inspector/internal/config"
	"github.com/rodweb/webhook-inspector/internal/fakerequests"
	"github.com/rodweb/webhook-inspector/internal/inspector"
	"github.com/rodweb/webhook-inspector/internal/ngrok"
	"github.com/rodweb/webhook-inspector/internal/request"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	EnvironmentVariableNgrokAuthToken = "NGROK_AUTHTOKEN"
	EnvironmentVariableNgrokDomain    = "NGROK_DOMAIN"
	EnvironmentVariableInspectorPort  = "INSPECTOR_PORT"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load .env file: %s\n", err)
	}

	cfg := config.Config{
		Token:  os.Getenv(EnvironmentVariableNgrokAuthToken),
		Domain: os.Getenv(EnvironmentVariableNgrokDomain),
		Port:   os.Getenv(EnvironmentVariableInspectorPort),
	}

	if err := run(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, config config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	shutdownFns := make([]func(), 0)

	errChan := make(chan error, 1)
	reqChan := make(chan request.Request)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	shutdownFns = append(shutdownFns, ngrok.Start(ctx, errChan, reqChan, ngrok.OptionsFromConfig(config)))
	shutdownFns = append(shutdownFns, inspector.Start(ctx, errChan, reqChan, inspector.OptionsFromConfig(config)))
	fakerequests.Start(reqChan)

	var wg sync.WaitGroup
	wg.Add(len(shutdownFns))

	select {
	case err := <-errChan:
		log.Printf("Received error, shutting down: %s\n", err)
		close(reqChan)
		cancel()
		return err
	case sgn := <-sigChan:
		log.Printf("Received signal, shutting down: %s\n", sgn)
		close(reqChan)
		cancel()
		for _, fn := range shutdownFns {
			go func(shutdown func()) {
				defer wg.Done()
				shutdown()
			}(fn)
		}
		wg.Wait()
	}

	log.Println("Gracefully shutdown done")
	return nil
}
