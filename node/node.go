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

func LoadNodesConfig(configFile string) (NodesInfo, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return NodesInfo{}, err
	}
	var nodesInfo NodesInfo
	if err := json.Unmarshal(file, &nodesInfo); err != nil {
		return NodesInfo{}, err
	}
	return nodesInfo, nil
}

func GetCurrentNodeId(config NodesInfo) (int32, error) {
	host, err := os.Hostname()
	if err != nil {
		return ErrNodeNotFound, err
	}
	for _, node := range config.Nodes {
		if node.Host == host {
			return node.Id, nil
		}
	}
	return ErrNodeNotFound, errors.New("current host not found in info")
}

func GetRandomNode(info NodesInfo) (Node, error) {
	if len(info.Nodes) == 0 {
		return Node{}, errors.New("no nodes available in info")
	}
	randIdx := seededRand.Intn(len(info.Nodes))
	return info.Nodes[randIdx], nil
}
