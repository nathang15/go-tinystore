// Representation of node in distributed system
package node

import (
	"encoding/json"
	"hash/crc32"
	"os"

	"github.com/nathang15/go-tinystore/pb"
)

type NodesInfo struct {
	Nodes map[string]*Node `json:"nodes"`
}

type Node struct {
	Id   string `json:"id"`
	Host string `json:"host"`
	Port int32  `json:"port"`
	// RestPort   int32  `json:"port"`
	// GrpcPort   int32  `json:"grpcPort"`
	HashId     uint32
	GrpcClient pb.CacheServiceClient
}

const (
	ErrNodeNotFound = -1
)

func InitNode(Id string) *Node {
	return &Node{
		Id:     Id,
		HashId: GetHashId(Id),
	}
}

func LoadNodesConfig(configFile string) NodesInfo {
	file, _ := os.ReadFile(configFile)
	nodesInfo := NodesInfo{}
	if err := json.Unmarshal([]byte(file), &nodesInfo); err != nil {
		return NodesInfo{}
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
	return ""
}

func GetHashId(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
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
