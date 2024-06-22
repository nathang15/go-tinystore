package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nathang15/go-tinystore/internal/ch"
	"github.com/nathang15/go-tinystore/internal/node"
	"github.com/nathang15/go-tinystore/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type Client struct {
	Info    node.NodesInfo
	Ring    *ch.Ring
	vNode   int
	CertDir string
}

type Payload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func InitClient(cert string, configFile string, virtualNodes int) *Client {
	initNodesConfig := node.LoadNodesConfig(configFile)
	ring := ch.InitRing(virtualNodes)
	var clusterConfig []*pb.Node

	for _, node := range initNodesConfig.Nodes {
		c, err := InitCacheClient(cert, node.Host, int(node.GrpcPort))
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		node.SetGrpcClient(c)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := c.GetClusterConfig(ctx, &pb.ClusterConfigRequest{CallerNodeId: "client"})
		if err != nil {
			log.Printf("error getting cluster config from node %s: %v", node.Id, err)
			continue
		}
		clusterConfig = res.Nodes
		break
	}

	infoMap := make(map[string]*node.Node)
	for _, n := range clusterConfig {
		infoMap[n.Id] = node.InitNode(n.Id, n.Host, n.RestPort, n.GrpcPort)

		if virtualNodes == 0 {
			ring.Add(n.Id, n.Host, n.RestPort, n.GrpcPort)
		} else {
			for i := 0; i < virtualNodes; i++ {
				virtualNodeID := fmt.Sprintf("%s-%d", n.Id, i)
				ring.Add(virtualNodeID, n.Host, n.RestPort, n.GrpcPort)
			}
		}
		c, err := InitCacheClient(cert, n.Host, int(n.GrpcPort))
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		infoMap[n.Id].SetGrpcClient(c)
	}
	info := node.NodesInfo{Nodes: infoMap}
	return &Client{Info: info, Ring: ring, vNode: virtualNodes, CertDir: cert}
}

func (c *Client) Get(key string) (string, error) {
	nodeId := c.Ring.Get(key)
	physicalNodeId := c.getPhysicalNodeId(nodeId)
	nodeInfo := c.Info.Nodes[physicalNodeId]

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/get", nodeInfo.Host, nodeInfo.RestPort))
	if err != nil {
		return "", fmt.Errorf("error sending GET request: %s", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("error reading response: %s", err)
	}

	return string(body), nil
}

func (c *Client) GetForGrpc(key string) (string, error) {
	nodeId := c.Ring.Get(key)
	physicalNodeId := c.getPhysicalNodeId(nodeId)
	nodeInfo := c.Info.Nodes[physicalNodeId]

	if nodeInfo.GrpcClient == nil {
		client, err := InitCacheClient(c.CertDir, nodeInfo.Host, int(nodeInfo.GrpcPort))
		if err != nil {
			return "", fmt.Errorf("error initiating gRPC client: %s", err)
		}
		nodeInfo.SetGrpcClient(client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := nodeInfo.GrpcClient.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		return "", fmt.Errorf("error gRPC GET: %s", err)
	}

	return res.GetData(), nil
}

func (c *Client) Put(key string, value string) error {
	nodeId := c.Ring.Get(key)
	physicalNodeId := c.getPhysicalNodeId(nodeId)
	if physicalNodeId == "" {
		return fmt.Errorf("no node found for key: %s", key)
	}
	nodeInfo, exists := c.Info.Nodes[physicalNodeId]
	if !exists {
		return fmt.Errorf("no node information for node ID: %s", physicalNodeId)
	}

	payload := Payload{Key: key, Value: value}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(payload)

	host := fmt.Sprintf("http://%s:%d/put", nodeInfo.Host, nodeInfo.RestPort)
	req, err := http.NewRequest("POST", host, b)
	if err != nil {
		return fmt.Errorf("error creating POST request: %s", err)
	}

	res, err := new(http.Client).Do(req)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("error sending POST request: %s", err)
	}
	return nil
}

func (client *Client) PutForGrpc(key string, value string) error {
	nodeId := client.Ring.Get(key)
	physicalNodeId := client.getPhysicalNodeId(nodeId)
	nodeInfo := client.Info.Nodes[physicalNodeId]

	if nodeInfo.GrpcClient == nil {
		client, err := InitCacheClient(client.CertDir, nodeInfo.Host, int(nodeInfo.GrpcPort))
		if err != nil {
			return fmt.Errorf("error initiating gRPC client: %s", err)
		}
		nodeInfo.SetGrpcClient(client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := nodeInfo.GrpcClient.Put(ctx, &pb.PutRequest{Key: key, Value: value})
	if err != nil {
		return fmt.Errorf("error making gRPC PUT: %s", err)
	}
	return nil
}

func InitCacheClient(cert string, server_host string, server_port int) (pb.CacheServiceClient, error) {
	creds, err := LoadTLSCredentials(cert)
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	var healthCheck = keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", server_host, server_port),
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(healthCheck),
		grpc.WithTimeout(time.Duration(time.Second)),
	)
	if err != nil {
		return nil, errors.New("failed to connect node")
	}

	return pb.NewCacheServiceClient(conn), nil
}

func LoadTLSCredentials(cert string) (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(fmt.Sprintf("%s/ca-cert.pem", cert))
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	clientCert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("%s/client-cert.pem", cert),
		fmt.Sprintf("%s/client-key.pem", cert),
	)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

func (c *Client) StartClusterConfigWatcher() {
	go func() {
		for {
			var leader *node.Node
			attempted := make(map[string]bool)
			for {
				randomNode := node.GetRandom(c.Ring.Nodes)

				if len(attempted) == len(c.Ring.Nodes) {
					log.Fatalf("Unable to connect to any nodes!")
				}

				if _, ok := attempted[randomNode.Id]; ok {
					log.Printf("Skipping visited node %s...", randomNode.Id)
					continue
				}

				attempted[randomNode.Id] = true

				client, err := InitCacheClient(c.CertDir, randomNode.Host, int(randomNode.GrpcPort))
				if err != nil {
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				res, err := client.GetLeader(ctx, &pb.LeaderRequest{Caller: "client"})
				if err != nil {
					continue
				}

				leader = c.Info.Nodes[res.Id]
				break
			}

			if leader == nil {
				continue
			}

			req := pb.ClusterConfigRequest{CallerNodeId: "client"}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client, err := InitCacheClient(c.CertDir, leader.Host, int(leader.GrpcPort))
			if err != nil {
				continue
			}

			res, err := client.GetClusterConfig(ctx, &req)
			if err != nil {
				log.Printf("Error getting cluster config from node %s at %s:%d: %v", leader.Id, leader.Host, leader.GrpcPort, err)
				continue
			}

			cluster_nodes := make(map[string]bool)
			for _, nodecfg := range res.Nodes {
				cluster_nodes[nodecfg.Id] = true
			}

			for _, node := range c.Info.Nodes {
				if _, ok := cluster_nodes[node.Id]; !ok {
					log.Printf("Removing node %s from ring", node.Id)
					delete(c.Info.Nodes, node.Id)
					c.Ring.Remove(node.Id)
				}
			}

			for _, nodeConfig := range res.Nodes {
				if _, ok := c.Info.Nodes[nodeConfig.Id]; !ok {
					log.Printf("Adding node %s to ring", nodeConfig.Id)
					c.Info.Nodes[nodeConfig.Id] = node.InitNode(nodeConfig.Id, nodeConfig.Host, nodeConfig.RestPort, nodeConfig.GrpcPort)
					c.Ring.Add(nodeConfig.Id, nodeConfig.Host, nodeConfig.RestPort, nodeConfig.GrpcPort)
				}
			}

			time.Sleep(3 * time.Second)
		}
	}()
}

func (c *Client) getPhysicalNodeId(virtualNodeId string) string {
	parts := strings.Split(virtualNodeId, "-")
	if len(parts) > 1 {
		return parts[0]
	}
	return virtualNodeId
}
