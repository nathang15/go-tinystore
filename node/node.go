// Representation of node in distributed system
package node

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"time"
)

type NodesInfo struct {
	Nodes []Node `json:"nodes"`
}

type Node struct {
	Id   int32  `json:"id"`
	Host string `json:"host"`
	Port int32  `json:"port"`
}

const (
	ErrNodeNotFound = -1
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func LoadNodesConfig(configFile string) NodesInfo {
	file, _ := os.ReadFile(configFile)
	nodesInfo := NodesInfo{}
	if err := json.Unmarshal([]byte(file), &nodesInfo); err != nil {
		return NodesInfo{}
	}
	return nodesInfo
}

func GetCurrentNodeId(config NodesInfo) int32 {
	host, _ := os.Hostname()

	for _, node := range config.Nodes {
		if node.Host == host {
			return node.Id
		}
	}
	return -1
}

func GetRandomNode(info NodesInfo) (Node, error) {
	if len(info.Nodes) == 0 {
		return Node{}, errors.New("no nodes available in info")
	}
	randIdx := seededRand.Intn(len(info.Nodes))
	return info.Nodes[randIdx], nil
}
