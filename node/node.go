// Representation of node in distributed system
package node

import (
	"encoding/json"
	"hash/crc32"
	"math/rand"
	"os"
	"time"

	"github.com/nathang15/go-tinystore/pb"
)

type NodesInfo struct {
	Nodes map[string]*Node `json:"nodes"`
}

type Node struct {
	Id         string `json:"id"`
	Host       string `json:"host"`
	RestPort   int32  `json:"port"`
	GrpcPort   int32  `json:"grpcPort"`
	HashId     uint32
	GrpcClient pb.CacheServiceClient
}

const (
	ErrNodeNotFound = -1
)

func InitNode(Id string, host string, restPort int32, grpcPort int32) *Node {
	return &Node{
		Id:       Id,
		Host:     host,
		RestPort: restPort,
		GrpcPort: grpcPort,
		HashId:   GetHashId(Id),
	}
}

func LoadNodesConfig(configFile string) NodesInfo {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return NodesInfo{}
	}

	var nodesInfo NodesInfo
	err = json.Unmarshal(file, &nodesInfo)
	if err != nil {
		return NodesInfo{}
	}

	if len(nodesInfo.Nodes) == 0 {
		nodesInfo.Nodes = make(map[string]*Node)
		defaultNode := InitNode("node0", "localhost", 8080, 5005)
		nodesInfo.Nodes[defaultNode.Id] = defaultNode
	} else {
		for _, nodeInfo := range nodesInfo.Nodes {
			nodeInfo.HashId = GetHashId(nodeInfo.Id)
		}
	}

	return nodesInfo
}

func GetCurrentNodeId(config NodesInfo) string {
	host, _ := os.Hostname()

	for _, node := range config.Nodes {
		if node.Host == host {
			return node.Id
		}
	}
	return randomSeq(5)
}

func GetRandom(nodes []*Node) *Node {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rand.Intn(len(nodes))
	return nodes[randomIndex]
}

func randomSeq(n int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetHashId(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (node *Node) SetGrpcClient(client pb.CacheServiceClient) {
	node.GrpcClient = client
}

type Nodes []*Node

func (n Nodes) Len() int {
	return len(n)
}
func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
func (n Nodes) Less(i, j int) bool {
	return n[i].HashId < n[j].HashId
}
