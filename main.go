package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/nathang15/go-tinystore/server"
)

func main() {
	gPort := flag.Int("grpc-port", 5005, "gRPC Server listens to this port")
	port := flag.Int("port", 8080, "HTTP Server listens to this port")
	capacity := flag.Int("capacity", 5, "LRU cache capacity")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	configFile := flag.String("config", "", "Path to nodes configuration file")
	flag.Parse()

	// TCP Connection
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *gPort))
	if err != nil {
		panic(err)
	}

	grpcServer, cacheServer := server.InitCacheServer(*capacity, *configFile, *verbose)
	log.Printf("Running gRPC server on port :%d", *gPort)
	go grpcServer.Serve(listener)
	cacheServer.RunElection()
	go cacheServer.MonitorLeaderStatus()
	cacheServer.RunHttpServer(*port)
}
