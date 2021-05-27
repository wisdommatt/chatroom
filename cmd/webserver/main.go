package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Meghee/kit/database/mongodb"
	"github.com/sirupsen/logrus"
	transportHTTP "github.com/wisdommatt/chatroom/internal/transport/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatal("An error occured while starting server !")
	}
}

func run() error {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})
	logger.SetReportCaller(true)
	logger.SetOutput(os.Stdout)

	mongoClient, err := mongodb.Connect(os.Getenv("MONGODB_DATABASE_URI"))
	if err != nil {
		logger.WithError(err).Panic("Database connection error")
		return err
	}
	defer mongoClient.Disconnect(context.TODO())
	mongoDB := mongoClient.Database(os.Getenv("DATABASE_NAME"))

	handler := transportHTTP.NewHandler(logger)
	handler.SetupRoutes(mongoDB)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3030"
	}

	server := http.Server{
		Addr:         ":" + port,
		Handler:      handler.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info("Server running on port: " + port)
		serverErrChan <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrChan:
		logger.WithError(err).Fatal("An error occured while starting server")
		return err

	case sig := <-shutdownChan:
		logger.WithField("signal", sig).Info("Gracefully shutting down server ...")
		shutdownServerCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownServerCtx); err != nil {
			server.Close()
			logger.WithError(err).Fatal("Graceful shutdown failed")
			return err
		}
		logger.Info("Graceful shutdown completed ....")
	}
	return nil
}
