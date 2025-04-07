package main

import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"worker/internal"
)

func main() {
	logger := &zap.Logger{}
	app := internal.NewApp(logger)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: http.HandlerFunc(app.PingHandler),
	}

	go func() {
		fmt.Println("HTTP server starting on port 8000...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	<-quit
	fmt.Println("\nShutting down server gracefully...")
	app.GracefulShutdown()
	srv.Close()
}
