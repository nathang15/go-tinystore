package server

import (
	"context"
	"fmt"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/nathang15/go-tinystore/node"
	"github.com/nathang15/go-tinystore/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CacheServer) RegisterNodeToCluster(ctx context.Context, nodeInfo *pb.Node) (*pb.GenericResponse, error) {
	if _, ok := s.nodesInfo.Nodes[nodeInfo.Id]; ok {
		s.logger.Infof("Node %s already part of cluster", nodeInfo.Id)
		return &pb.GenericResponse{Data: SUCCESS}, nil
	}

	s.nodesInfo.Nodes[nodeInfo.Id] = node.InitNode(nodeInfo.Id, nodeInfo.Host, nodeInfo.RestPort, nodeInfo.GrpcPort)

	var nodes []*pb.Node
	for _, node := range s.nodesInfo.Nodes {
		nodes = append(nodes, &pb.Node{Id: node.Id, Host: node.Host, RestPort: node.RestPort, GrpcPort: node.GrpcPort})
	}
	for _, node := range s.nodesInfo.Nodes {
		if node.Id == s.nodeId {
			continue
		}
		s.logger.Infof("Sending updated cluster config to node %s with grpc client %v", node.Id, node.GrpcClient)
		req_ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		cfg := pb.ClusterConfig{Nodes: nodes}

		c, err := s.ServerInitCacheClient(nodeInfo.Host, int(nodeInfo.GrpcPort))
		if err != nil {
			s.logger.Errorf("unable to connect to node %s", nodeInfo.Id)
			return nil, status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("Unable to connect to node being registered: %s", nodeInfo.Id),
			)
		}
		c.UpdateClusterConfig(req_ctx, &cfg)
	}
	return &pb.GenericResponse{Data: SUCCESS}, nil
}

func (s *CacheServer) GetClusterConfig(ctx context.Context, req *pb.ClusterConfigRequest) (*pb.ClusterConfig, error) {
	var nodes []*pb.Node
	for _, node := range s.nodesInfo.Nodes {
		nodes = append(nodes, &pb.Node{Id: node.Id, Host: node.Host, RestPort: node.RestPort, GrpcPort: node.GrpcPort})
	}
	s.logger.Infof("Returning cluster config to node %s: %v", req.CallerNodeId, nodes)
	return &pb.ClusterConfig{Nodes: nodes}, nil
}

func (s *CacheServer) UpdateClusterConfig(ctx context.Context, req *pb.ClusterConfig) (*empty.Empty, error) {
	s.logger.Info("Updating cluster config")
	s.nodesInfo.Nodes = make(map[string]*node.Node)
	for _, nodecfg := range req.Nodes {
		s.nodesInfo.Nodes[nodecfg.Id] = node.InitNode(nodecfg.Id, nodecfg.Host, nodecfg.RestPort, nodecfg.GrpcPort)
	}
	return &empty.Empty{}, nil
}

func (s *CacheServer) updateClusterConfigInternal() {
	s.logger.Info("Sending out updated cluster config")
	var nodes []*pb.Node
	for _, node := range s.nodesInfo.Nodes {
		nodes = append(nodes, &pb.Node{Id: node.Id, Host: node.Host, RestPort: node.RestPort, GrpcPort: node.GrpcPort})
	}
	for _, node := range s.nodesInfo.Nodes {
		if node.Id == s.nodeId {
			continue
		}
		reqCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		cfg := pb.ClusterConfig{Nodes: nodes}

		c, err := s.ServerInitCacheClient(node.Host, int(node.GrpcPort))

		if err != nil {
			s.logger.Errorf("unable to connect to node %s", node.Id)
			continue
		}

		_, err = c.UpdateClusterConfig(reqCtx, &cfg)
		if err != nil {
			s.logger.Infof("error sending cluster config to node %s: %v", node.Id, err)
		}
	}
}
