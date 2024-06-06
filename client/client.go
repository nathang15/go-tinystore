package client

import (
	"github.com/nathang15/go-tinystore/node"
	"github.com/nathang15/go-tinystore/ring"
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
	r := ring.InitRing(virtualNodes)
	for _, node := range nodesInfo.Nodes {
		r.Add(node.Id, node.Host, node.Port)
	}
	return &Client{Info: nodesInfo, Ring: r, vNode: virtualNodes}
}

//TODO: GET

//TODO: PUT
