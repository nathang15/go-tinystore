package server

import (
	"context"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/nathang15/go-tinystore/node"
	"github.com/nathang15/go-tinystore/pb"
)

func (s *CacheServer) AddNodeToCluster(ctx context.Context, nodeInfo *pb.Node) (*pb.GenericResponse, error) {
	if _, ok := s.nodesInfo.Nodes[nodeInfo.Id]; ok {
		s.logger.Infof("Node %s is already part of the cluster", nodeInfo.Id)
		return &pb.GenericResponse{Data: SUCCESS}, nil
	}

	s.Ring.Add(nodeInfo.Id, nodeInfo.Host, nodeInfo.RestPort, nodeInfo.GrpcPort)
	s.logger.Infof("Node %s added to the ring with host %s", nodeInfo.Id, nodeInfo.Host)

	s.nodesInfo.Nodes[nodeInfo.Id] = node.InitNode(nodeInfo.Id, nodeInfo.Host, nodeInfo.RestPort, nodeInfo.GrpcPort)
	s.logger.Infof("Node %s initialized and added to the cluster", nodeInfo.Id)

	s.updateClusterConfigInternal()

	return &pb.GenericResponse{Data: SUCCESS}, nil
}

func (s *CacheServer) GetClusterConfig(ctx context.Context, req *pb.ClusterConfigRequest) (*pb.ClusterConfig, error) {
	s.logger.Infof("Received request to get cluster config from node %s", req.CallerNodeId)
	nodes := s.getClusterNodes()
	s.logger.Infof("Cluster config for node %s: %v", req.CallerNodeId, nodes)
	return &pb.ClusterConfig{Nodes: nodes}, nil
}

func (s *CacheServer) UpdateClusterConfig(ctx context.Context, req *pb.ClusterConfig) (*empty.Empty, error) {
	s.logger.Info("Updating cluster configuration")
	for _, nodecfg := range req.Nodes {
		if _, ok := s.nodesInfo.Nodes[nodecfg.Id]; !ok {
			s.Ring.Add(nodecfg.Id, nodecfg.Host, nodecfg.RestPort, nodecfg.GrpcPort)
			s.nodesInfo.Nodes[nodecfg.Id] = node.InitNode(nodecfg.Id, nodecfg.Host, nodecfg.RestPort, nodecfg.GrpcPort)
			s.logger.Infof("New node %s added to the ring with host %s", nodecfg.Id, nodecfg.Host)
		}
	}
	return &empty.Empty{}, nil
}

func (s *CacheServer) updateClusterConfigInternal() {
	s.logger.Info("Sending out updated cluster config")

	var nodes []*pb.Node
	for _, ringnode := range s.Ring.Nodes {
		nodes = append(nodes, &pb.Node{Id: ringnode.Id, Host: ringnode.Host, RestPort: ringnode.RestPort, GrpcPort: ringnode.GrpcPort})
	}
	for _, ringnode := range s.Ring.Nodes {
		req_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		cfg := pb.ClusterConfig{Nodes: nodes}

		ringnode.GrpcClient.UpdateClusterConfig(req_ctx, &cfg)
	}
}

// helper function to get cluster nodes
func (s *CacheServer) getClusterNodes() []*pb.Node {
	var nodes []*pb.Node
	for _, rnode := range s.Ring.Nodes {
		nodes = append(nodes, &pb.Node{Id: rnode.Id, Host: rnode.Host, RestPort: rnode.RestPort, GrpcPort: rnode.GrpcPort})
	}
	return nodes
}
