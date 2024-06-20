package server

import (
	"context"
	"os"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/nathang15/go-tinystore/pb"
)

const (
	LEADER      = "LEADER"
	FOLLOWER    = "FOLLOWER"
	RUNNING     = true
	NO_ELECTION = false
	NO_LEADER   = "NO LEADER"
)

// Leader Election with bully algorithm
func (s *CacheServer) RunElection() {
	if s.electionStatus {
		s.logger.Info("Election already running")
		return
	}

	s.electionStatus = RUNNING

	pid := int32(os.Getpid())
	s.logger.Infof("Starting election with pid %d", pid)

	for _, node := range s.nodesInfo.Nodes {
		if node.Id == s.nodeId {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		c, err := s.ServerInitCacheClient(node.Host, int(node.GrpcPort))
		if err != nil {
			s.logger.Infof("error creating grpc client to node %s: %v", node.Id, err)
		}
		res, err := c.GetPid(ctx, &pb.PidRequest{CallerPid: pid})

		if err != nil {
			s.logger.Infof("Error getting pid from node %d: %v", node.Id, err)
			continue
		}

		s.logger.Infof("Got pid %d from node %d", res.Pid, node.Id)
		if (pid < res.Pid) || (res.Pid == pid && s.nodeId < node.Id) {

			s.logger.Infof("Sending election request to node %s", node.Id)

			c, err := s.ServerInitCacheClient(node.Host, int(node.GrpcPort))
			if err != nil {
				s.logger.Infof("error creating grpc client to node node %s: %v", node.Id, err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			_, err = c.RequestElection(ctx, &pb.ElectionRequest{CallerPid: pid, CallerNodeId: s.nodeId})
			if err != nil {
				s.logger.Infof("Error requesting node %s run an election: %v", node.Id, err)
			}

			s.logger.Info("Waiting for decision")
			select {
			case s.leaderId = <-s.decisionChannel:
				s.logger.Infof("Received decision: Leader is node %s", s.leaderId)
				s.electionStatus = NO_ELECTION
				return
			case <-time.After(5 * time.Second):
				s.logger.Info("Timed out waiting for decision. Starting new election.")
				s.RunElection()
				s.electionStatus = NO_ELECTION
				return
			}
		}
	}

	s.leaderId = s.nodeId

	s.SetNewLeader(s.leaderId)

	s.electionStatus = NO_ELECTION

}

// Notify new leader to all nodes
func (s *CacheServer) SetNewLeader(newLeader string) {
	s.logger.Infof("Announcing node %s won election", newLeader)

	for _, node := range s.nodesInfo.Nodes {
		if node.Id == s.nodeId {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		c, err := s.ServerInitCacheClient(node.Host, int(node.GrpcPort))
		if err != nil {
			s.logger.Infof("error creating grpc client to node node %s: %v", node.Id, err)
		}

		_, err = c.UpdateLeader(ctx, &pb.NewLeaderAnnouncement{LeaderId: newLeader})

		if err != nil {
			s.logger.Infof("Election leader announcement to node %s error: %v", node.Id, err)
		}
		cancel()
	}
}

func (s *CacheServer) GetLeader(ctx context.Context, request *pb.LeaderRequest) (*pb.LeaderResponse, error) {
	for {
		if s.leaderId != NO_LEADER {
			break
		}
		s.RunElection()

		if s.leaderId == NO_LEADER {
			s.logger.Info("No leader elected, waiting 3 seconds before trying again...")
			time.Sleep(3 * time.Second)
		}
	}
	return &pb.LeaderResponse{Id: s.leaderId}, nil
}

// leader status monitoring
func (s *CacheServer) MonitorLeaderStatus() {
	s.logger.Info("Leader status monitor starting...")

	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C

		if s.leaderId != s.nodeId {
			if !s.IsLeaderUp() {
				s.logger.Info("Leader status lost, running new election")
				s.RunElection()
				s.logger.Info("Selected new leader! Continue monitoring")
			}
			select {
			case <-s.shutdownChannel:
				s.logger.Info("Received shutdown signal")
				break
			case <-time.After(time.Second):
				continue
			}

		} else {
			modified := false
			for _, node := range s.nodesInfo.Nodes {
				if node.Id == s.nodeId {
					continue
				}

				client, err := s.ServerInitCacheClient(node.Host, int(node.GrpcPort))
				if err != nil {
					s.logger.Infof("error creating grpc client to node node %s: %v", node.Id, err)
					delete(s.nodesInfo.Nodes, node.Id)
					modified = true
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				s.logger.Infof("Checking status of node %s", node.Id)

				_, err = client.GetStatus(ctx, &pb.StatusRequest{CallerNodeId: s.nodeId})

				if err != nil {
					s.logger.Infof("Node %s status returned error, removing from cluster", node.Id, err)
					delete(s.nodesInfo.Nodes, node.Id)
					modified = true
				}
			}

			if modified {
				s.logger.Info("Detected node config change, sending update to other nodes")
				s.updateClusterConfigInternal()
			}
		}
	}
}

func (s *CacheServer) GetStatus(ctx context.Context, req *pb.StatusRequest) (*empty.Empty, error) {
	s.logger.Infof("Node %s returning status to node %s", s.nodeId, req.CallerNodeId)
	return &empty.Empty{}, nil
}

func (s *CacheServer) IsLeaderUp() bool {
	if s.leaderId == NO_LEADER {
		s.logger.Info("No leader existed!")
		return false
	}

	if s.nodeId == s.leaderId {
		return true
	}

	s.logger.Infof("leader is %s", s.leaderId)

	leader, exists := s.nodesInfo.Nodes[s.leaderId]
	if !exists {
		s.logger.Infof("Leader node %s not found in nodesInfo", s.leaderId)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := s.ServerInitCacheClient(leader.Host, int(leader.GrpcPort))
	if err != nil {
		s.logger.Infof("error creating grpc client to node %s: %v", leader.Id, err)
		return false
	}

	_, err = c.GetStatus(ctx, &pb.StatusRequest{CallerNodeId: s.nodeId})

	if err != nil {
		s.logger.Infof("Leader status check error: %v", err)
		return false
	}

	return true
}

func (s *CacheServer) UpdateLeader(ctx context.Context, req *pb.NewLeaderAnnouncement) (*pb.GenericResponse, error) {
	s.logger.Infof("Received announcement leader is %s", req.LeaderId)
	s.leaderId = req.LeaderId
	s.decisionChannel <- s.leaderId
	return &pb.GenericResponse{Data: SUCCESS}, nil
}

func (s *CacheServer) GetPid(ctx context.Context, req *pb.PidRequest) (*pb.PidResponse, error) {
	localPid := int32(os.Getpid())
	return &pb.PidResponse{Pid: localPid}, nil
}

func (s *CacheServer) RequestElection(ctx context.Context, req *pb.ElectionRequest) (*pb.GenericResponse, error) {
	s.logger.Infof("received request for election from %s", req.CallerNodeId)
	go s.RunElection()
	return &pb.GenericResponse{Data: SUCCESS}, nil
}
