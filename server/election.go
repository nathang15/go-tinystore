package server

import (
	"context"
	"os"
	"time"

	"github.com/nathang15/go-tinystore/pb"
)

const (
	LEADER      = "LEADER"
	FOLLOWER    = "FOLLOWER"
	RUNNING     = true
	NO_ELECTION = false
	NO_LEADER   = "NO LEADER"
	SUCCESS     = "OK"
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

		client := NewGrpcClientForNode(node)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := client.GetPid(ctx, &pb.PidRequest{CallerPid: pid})
		if err != nil {
			s.logger.Infof("Error getting pid from node %d: %v", node.Id, err)
			continue
		}

		//if resposne has higher pid -> election request -> wait for leader result
		s.logger.Infof("Got pid %d from node %d", res.Pid, node.Id)
		if (pid < res.Pid) || (res.Pid == pid && s.nodeId < node.Id) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			s.logger.Infof("Sending election request to node %d", node.Id)
			_, err = client.RequestElection(ctx, &pb.ElectionRequest{CallerPid: pid, CallerNodeId: s.nodeId})
			if err != nil {
				s.logger.Infof("Error sending election request to node %d: %v", node.Id, err)
			}
		}

		s.logger.Infof("Electing leader!")
		select {
		case s.leaderId = <-s.decisionChannel:
			s.logger.Infof("Leader elected: %d", s.leaderId)
			s.electionStatus = NO_ELECTION
			return
		case <-time.After(5 * time.Second):
			s.logger.Infof("No leader elected")
			s.RunElection()
			s.electionStatus = NO_ELECTION
			return
		}
	}

	s.leaderId = s.nodeId

	s.SetNewLeader(s.leaderId)

	s.electionStatus = NO_ELECTION

}

// Notify new leader to all nodes
func (s *CacheServer) SetNewLeader(newLeader string) {
	s.logger.Infof("New leader elected: %d", newLeader)

	for _, node := range s.nodesInfo.Nodes {
		if node.Id == s.nodeId {
			continue
		}

		client := NewGrpcClientForNode(node)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.UpdateLeader(ctx, &pb.NewLeaderAnnouncement{LeaderId: newLeader})
		if err != nil {
			s.logger.Infof("Error updating leader to node %d: %v", node.Id, err)
			continue
		}
	}
}

// leader status monitoring
func (s *CacheServer) MonitorLeaderStatus() {
	s.logger.Info("Monitoring leader status")

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if !s.IsLeaderUp() {
				s.logger.Info("Leader is down, running election")
				s.RunElection()
				s.logger.Info("Selected new leader!")
			}
		case <-s.shutdownChannel:
			s.logger.Info("Shutting down leader status monitoring")
			return
		}
	}
}

func (s *CacheServer) IsLeaderUp() bool {
	if s.leaderId == NO_LEADER {
		s.logger.Info("No leader existed!")
		return false
	}

	if s.nodeId == s.leaderId {
		return true
	}

	leader := s.nodesInfo.Nodes[s.leaderId]

	client := NewGrpcClientForNode(leader)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetStatus(ctx, &pb.StatusRequest{CallerNodeId: s.nodeId})
	if err != nil {
		s.logger.Infof("Leader %d is down: %v", s.leaderId, err)
		return false
	}
	return true
}

func (s *CacheServer) UpdateLeader(ctx context.Context, req *pb.NewLeaderAnnouncement) (*pb.GenericResponse, error) {
	s.leaderId = req.LeaderId
	s.decisionChannel <- s.leaderId
	return &pb.GenericResponse{Data: SUCCESS}, nil
}

func (s *CacheServer) GetStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	var status string
	if s.nodeId == s.leaderId {
		status = LEADER
	} else {
		status = FOLLOWER
	}

	s.logger.Infof("Node %d returning status %s to node %d", s.nodeId, status, req.CallerNodeId)
	return &pb.StatusResponse{Status: status, NodeId: s.nodeId}, nil
}

func (s *CacheServer) GetPid(ctx context.Context, req *pb.PidRequest) (*pb.PidResponse, error) {
	localPid := int32(os.Getpid())
	if localPid > req.CallerPid {
		go s.RunElection()
	}
	return &pb.PidResponse{Pid: localPid}, nil
}

func (s *CacheServer) RequestElection(ctx context.Context, req *pb.ElectionRequest) (*pb.GenericResponse, error) {
	if s.electionStatus {
		return &pb.GenericResponse{Data: "Election already running"}, nil
	} else {
		go s.RunElection()
	}
	return &pb.GenericResponse{Data: SUCCESS}, nil
}
