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
	"time"
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

type ShutdownFunc func(ctx context.Context)

func run(ctx context.Context, config config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	shutdownFns := make([]ShutdownFunc, 0)

	errChan := make(chan error, 1)
	reqChan := make(chan request.Request)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	shutdownFns = append(shutdownFns, ngrok.Start(ctx, errChan, reqChan, ngrok.OptionsFromConfig(config)))
	shutdownFns = append(shutdownFns, inspector.Start(ctx, errChan, reqChan, inspector.OptionsFromConfig(config)))
	fakerequests.Start(ctx, reqChan)

	var err error
	select {
	case e := <-errChan:
		log.Printf("Received error, shutting down: %s\n", err)
		err = e
	case sgn := <-sigChan:
		log.Printf("Received signal, shutting down: %s\n", sgn)
	}

	cancel()
	close(reqChan)

	var wg sync.WaitGroup
	wg.Add(len(shutdownFns))

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, fn := range shutdownFns {
		go func(shutdown ShutdownFunc) {
			defer wg.Done()
			shutdown(ctxTimeout)
		}(fn)
	}
	wg.Wait()

	log.Println("Gracefully shutdown done")
	time.Sleep(1 * time.Nanosecond)
	return err
}
