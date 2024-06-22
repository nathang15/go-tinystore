package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nathang15/go-tinystore/internal/server"
)

func main() {
	grpc_port := flag.Int("grpc-port", 5005, "port number for gRPC server")
	capacity := flag.Int("capacity", 5, "lru capacity")
	verbose := flag.Bool("verbose", false, "events log")
	config_file := flag.String("config", "", "JSON config file")
	rest_port := flag.Int("rest-port", 8080, "enable REST API for client requests too")

	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpc_port))
	if err != nil {
		panic(err)
	}

	grpc_server, cache_server := server.InitCacheServer(*capacity, *config_file, *verbose, server.DYNAMIC)

	log.Printf("Running gRPC server on: %d", *grpc_port)
	go grpc_server.Serve(listener)

	cache_server.RegisterNodeInternal()

	cache_server.RunElection()

	go cache_server.MonitorLeaderStatus()

	log.Printf("Running REST API server on: %d", *rest_port)
	http_server := cache_server.RunHttpServer(*rest_port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c

		log.Printf("Shutting down gRPC server!")
		grpc_server.Stop()

		log.Printf("Shutting down HTTP server!")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := http_server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %s", err)
		}
		os.Exit(0)
	}()

	select {}

	// ring.PrintBucketDistribution()
}
