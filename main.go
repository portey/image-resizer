package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/portey/image-resizer/graph"
	"github.com/portey/image-resizer/graph/resolver"
	"github.com/portey/image-resizer/healthcheck"
	"github.com/portey/image-resizer/opts"
	"github.com/portey/image-resizer/repository/mongo"
	"github.com/portey/image-resizer/resizer"
	"github.com/portey/image-resizer/service"
	"github.com/portey/image-resizer/storage/minio"
	log "github.com/sirupsen/logrus"
)

func main() {
	config := opts.ReadOS()
	initLogger(config.LogLevel, config.PrettyLogOutput)

	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)

	storage, err := minio.New(config.StorageCfg)
	if err != nil {
		log.Fatalf("storage initialization %v", err)
	}

	repo, err := mongo.New(ctx, config.MongoURI, config.MongoDatabase)
	if err != nil {
		log.Fatalf("repository initialization %v", err)
	}

	srv := service.New(storage, resizer.New(), repo)

	graphqlResolver := resolver.New(srv)
	graphqlSrv := graph.New(config.GraphQLPort, graphqlResolver)

	healthCheckSrv := healthcheck.New(config.HealthCHeckPort, []healthcheck.Check{
		repo.Ping,
		graphqlSrv.HealthCheck,
	})

	var wg sync.WaitGroup
	graphqlSrv.Run(ctx, &wg)
	healthCheckSrv.Run(ctx, &wg)
	wg.Wait()
}

func setupGracefulShutdown(stop func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Println("Got Interrupt signal")
		stop()
	}()
}

func initLogger(logLevel string, pretty bool) {
	if pretty {
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetOutput(os.Stderr)

	switch strings.ToLower(logLevel) {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}
