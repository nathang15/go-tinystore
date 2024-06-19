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

		res, err := node.GrpcClient.GetPid(ctx, &pb.PidRequest{CallerPid: pid})
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
			_, err = node.GrpcClient.RequestElection(ctx, &pb.ElectionRequest{CallerPid: pid, CallerNodeId: s.nodeId})
			if err != nil {
				s.logger.Infof("Error sending election request to node %d: %v", node.Id, err)
			}
		}

		s.logger.Infof("Electing leader!")
		select {
		case s.leaderId = <-s.decisionChannel:
			s.logger.Infof("Received decision: Leader is node %s", s.leaderId)
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
	s.logger.Infof("Announcing node %s won election", newLeader)

	for _, node := range s.nodesInfo.Nodes {
		if node.Id == s.nodeId {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := node.GrpcClient.UpdateLeader(ctx, &pb.NewLeaderAnnouncement{LeaderId: newLeader})
		if err != nil {
			s.logger.Infof("Election leader announcement to node %s error: %v", node.Id, err)
			continue
		}
	}

	s.leaderId = newLeader
	s.logger.Infof("New leader set: %s", s.leaderId)
}

// Returns current leader
func (s *CacheServer) GetLeader(ctx context.Context, request *pb.LeaderRequest) (*pb.LeaderResponse, error) {
	// while there is no leader, run election
	for {
		if s.leaderId != NO_LEADER {
			break
		}
		s.RunElection()

		// if no leader was elected, wait 3 seconds then run another election
		if s.leaderId == NO_LEADER {
			s.logger.Info("No leader elected, waiting 3 seconds before trying again...")
			time.Sleep(3 * time.Second)
		}
	}
	return &pb.LeaderResponse{Id: s.leaderId}, nil
}

// leader status monitoring
func (s *CacheServer) MonitorLeaderStatus() {
	s.logger.Info("Leader heartbeat monitor starting...")

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
			for _, node := range s.Ring.Nodes {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				_, err := node.GrpcClient.GetStatus(ctx, &pb.StatusRequest{CallerNodeId: s.nodeId})
				if err != nil {
					s.logger.Infof("Node %s healthcheck returned error, removing from cluster", node.Id, err)
					delete(s.nodesInfo.Nodes, node.Id)
					s.Ring.Remove(node.Id)
					modified = true
				}
			}

			if modified {
				s.updateClusterConfigInternal()
			}
		}
	}
}

func (s *CacheServer) GetStatus(ctx context.Context, req *pb.StatusRequest) (*empty.Empty, error) {
	var status string
	if s.nodeId == s.leaderId {
		status = LEADER
	} else {
		status = FOLLOWER
	}

	s.logger.Infof("Node %s returning status %s to node %s", s.nodeId, status, req.CallerNodeId)
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

	leader, exists := s.nodesInfo.Nodes[s.leaderId]
	if !exists {
		s.logger.Infof("Leader node %s not found in nodesInfo", s.leaderId)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := leader.GrpcClient.GetStatus(ctx, &pb.StatusRequest{CallerNodeId: s.nodeId})
	if err != nil {
		s.logger.Infof("Leader %s is down: %v", s.leaderId, err)
		return false
	}

	return true
}

func (s *CacheServer) UpdateLeader(ctx context.Context, req *pb.NewLeaderAnnouncement) (*pb.GenericResponse, error) {
	s.leaderId = req.LeaderId
	s.decisionChannel <- s.leaderId
	return &pb.GenericResponse{Data: SUCCESS}, nil
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
