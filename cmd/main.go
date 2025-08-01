package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/smncd/squishy/internal/filesystem"
	"gitlab.com/smncd/squishy/internal/logging"
	"gitlab.com/smncd/squishy/internal/server"
)

var Version string

func main() {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	s := &filesystem.SquishyFile{}

	s.SetFilePath("squishy.yaml")

	err := s.Load()
	if err != nil {
		logging.Error(logger, "Error loading config: %v", err)
		os.Exit(1)
	}

	server := server.New(s, logger)

	logger.Printf("Starting Squishy v%s...", Version)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error(logger, "listen: %s\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutdown Squishy ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalln("Server Shutdown:", err)
	}

	logger.Println("Squishy exiting")
}
