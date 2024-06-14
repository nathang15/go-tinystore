package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nathang15/go-tinystore/node"
	"github.com/nathang15/go-tinystore/pb"
	"github.com/nathang15/go-tinystore/ring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	Info  node.NodesInfo
	Ring  *ring.Ring
	vNode int
}

type Payload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func InitClient(configFile string, virtualNodes int) *Client {
	nodesInfo := node.LoadNodesConfig(configFile)
	if len(nodesInfo.Nodes) == 0 {
		log.Fatal("No nodes found in configuration")
	}
	r := ring.InitRing(virtualNodes)
	for _, node := range nodesInfo.Nodes {
		r.Add(node.Id, node.Host, node.RestPort, node.GrpcPort)
	}
	return &Client{Info: nodesInfo, Ring: r, vNode: virtualNodes}
}

func (c *Client) Get(key string) string {
	nodeId := c.Ring.Get(key)
	nodeInfo := c.Info.Nodes[nodeId]

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/get", nodeInfo.Host, nodeInfo.RestPort))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func (c *Client) GetForGrpc(key string) {
	nodeId := c.Ring.Get(key)
	nodeInfo := c.Info.Nodes[nodeId]

	client := InitCacheClient(nodeInfo.Host, int(nodeInfo.GrpcPort))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		log.Fatalf("Error getting key %s: %v", key, err)
		return
	}

	log.Printf("Got value %s for key %s", res.GetData(), key)
}

func (c *Client) Put(key string, value string) (string, error) {
	nodeId := c.Ring.Get(key)
	if nodeId == "" {
		return "", fmt.Errorf("no node found for key: %s", key)
	}
	nodeInfo, exists := c.Info.Nodes[nodeId]
	if !exists {
		return "", fmt.Errorf("no node information for node ID: %s", nodeId)
	}

	payload := Payload{Key: key, Value: value}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(payload)

	host := fmt.Sprintf("http://%s:%d/put", nodeInfo.Host, nodeInfo.RestPort)
	req, err := http.NewRequest("POST", host, b)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *Client) PutForGrpc(key string, value string) {
	nodeId := c.Ring.Get(key)
	if nodeId == "" {
		return
	}
	nodeInfo, exists := c.Info.Nodes[nodeId]
	if !exists {
		return
	}

	client := InitCacheClient(nodeInfo.Host, int(nodeInfo.GrpcPort))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Put(ctx, &pb.PutRequest{Key: key, Value: value})
	if err != nil {
		log.Fatalf("Error putting key '%s' value '%s' into cache: %v", key, value, err)
		return
	}
}

func InitCacheClient(server_host string, server_port int) pb.CacheServiceClient {
	creds, err := LoadTLSCredentials()
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", server_host, server_port), grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}

	return pb.NewCacheServiceClient(conn)
}

func LoadTLSCredentials() (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile("../certs/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	clientCert, err := tls.LoadX509KeyPair("../certs/client-cert.pem", "../certs/client-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
