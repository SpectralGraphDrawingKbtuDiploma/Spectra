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

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	logger := &zap.Logger{}
	app := internal.NewApp(logger)
	// Channel to listen for signals (SIGINT, SIGTERM)
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
