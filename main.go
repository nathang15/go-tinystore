package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/nathang15/go-tinystore/server"
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

	grpc_server, cache_server := server.InitCacheServer(*capacity, *config_file, *verbose)

	log.Printf("Running gRPC server on: %d", *grpc_port)
	go grpc_server.Serve(listener)

	// cache_server.SetAllGrpcClients()

	cache_server.RegisterNodeInternal()

	cache_server.RunElection()

	go cache_server.MonitorLeaderStatus()

	log.Printf("Running REST API server on: %d", *rest_port)
	cache_server.RunHttpServer(*rest_port)
}

// OUTPUT DISTRIBUTT
// func main() {
// 	ring.PrintBucketDistribution()
// }
