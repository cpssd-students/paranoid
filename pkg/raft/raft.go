// Package raft stores function used to interface in and out of raft.
package raft

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	"github.com/cpssd-students/paranoid/pkg/raft/raftlog"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

// Raft constants
const (
	ElectionTimeout      time.Duration = 3000 * time.Millisecond
	Heartbeat                          = 1000 * time.Millisecond
	RequestVoteTimeout                 = 5500 * time.Millisecond
	HeartbeatTimeout                   = 3000 * time.Millisecond
	SendEntryTimeout                   = 7500 * time.Millisecond
	EntryAppliedTimeout                = 20000 * time.Millisecond
	leaderRequestTimeout               = 10000 * time.Millisecond
)

// MaxAppendEntries that can be send it one append request
const MaxAppendEntries uint64 = 100

var _ pb.RaftNetworkServiceServer = (*NetworkServer)(nil)

// NetworkServer implements the raft protobuf server interface
type NetworkServer struct {
	pb.UnimplementedRaftNetworkServiceServer

	State *State
	Wait  sync.WaitGroup

	nodeDetails       Node
	raftInfoDirectory string
	TLSEnabled        bool
	Encrypted         bool
	TLSSkipVerify     bool

	QuitChannelClosed    bool
	Quit                 chan bool
	ElectionTimeoutReset chan bool

	appendEntriesLock sync.Mutex
	addEntryLock      sync.Mutex
}

// AppendEntries implementation
func (s *NetworkServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	s.appendEntriesLock.Lock()
	defer s.appendEntriesLock.Unlock()

	if !s.State.Configuration.InConfiguration(req.LeaderId) {
		if s.State.Configuration.MyConfigurationGood() {
			return &pb.AppendEntriesResponse{
				Term:    s.State.GetCurrentTerm(),
				Success: false,
			}, nil
		}
	}

	if req.Term < s.State.GetCurrentTerm() {
		return &pb.AppendEntriesResponse{
			Term:      s.State.GetCurrentTerm(),
			NextIndex: 0,
			Success:   false,
		}, nil
	}

	s.ElectionTimeoutReset <- true
	s.State.SetLeaderID(req.LeaderId)

	if req.Term > s.State.GetCurrentTerm() {
		s.State.SetCurrentTerm(req.Term)
		s.State.SetCurrentState(FOLLOWER)
	}

	if req.PrevLogIndex != 0 {
		if s.State.Log.GetMostRecentIndex() < req.PrevLogIndex {
			return &pb.AppendEntriesResponse{
				Term:      s.State.GetCurrentTerm(),
				NextIndex: s.State.Log.GetMostRecentIndex() + 1,
				Success:   false,
			}, nil
		}
		preLogEntry, err := s.State.Log.GetLogEntry(req.PrevLogIndex)
		if err != nil && err != raftlog.ErrIndexBelowStartIndex {
			log.Fatalf("Unable to get log entry: %v", err)
		} else if err == raftlog.ErrIndexBelowStartIndex {
			if req.PrevLogIndex != s.State.Log.GetMostRecentIndex() {
				return &pb.AppendEntriesResponse{
					Term:      s.State.GetCurrentTerm(),
					NextIndex: 0,
					Success:   false,
				}, nil
			}
			if req.PrevLogTerm != s.State.Log.GetMostRecentTerm() {
				return &pb.AppendEntriesResponse{
					Term:      s.State.GetCurrentTerm(),
					NextIndex: 0,
					Success:   false,
				}, nil
			}
		} else if err == nil {
			if preLogEntry.Term != req.PrevLogTerm {
				return &pb.AppendEntriesResponse{
					Term:      s.State.GetCurrentTerm(),
					NextIndex: 0,
					Success:   false,
				}, nil
			}
		}
	}

	for i := uint64(0); i < uint64(len(req.Entries)); i++ {
		logIndex := req.PrevLogIndex + 1 + i

		if s.State.Log.GetMostRecentIndex() >= logIndex {
			logEntryAtIndex, err := s.State.Log.GetLogEntry(logIndex)
			if err != nil && err != raftlog.ErrIndexBelowStartIndex {
				log.Fatalf("Unable to get log entry: %v", err)
			} else if err == nil {
				if logEntryAtIndex.Term != req.Term {
					_ = s.State.Log.DiscardLogEntriesAfter(logIndex)
					s.appendLogEntry(req.Entries[i])
				}
			}
		} else {
			s.appendLogEntry(req.Entries[i])
		}
	}

	if req.LeaderCommit > s.State.GetCommitIndex() {
		lastLogIndex := s.State.Log.GetMostRecentIndex()
		if lastLogIndex < req.LeaderCommit {
			s.State.SetCommitIndex(lastLogIndex)
		} else {
			s.State.SetCommitIndex(req.LeaderCommit)
		}
	}

	return &pb.AppendEntriesResponse{
		Term:      s.State.GetCurrentTerm(),
		NextIndex: 0,
		Success:   true,
	}, nil
}

// RequestVote implementation
func (s *NetworkServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	if !s.State.Configuration.InConfiguration(req.CandidateId) {
		if s.State.Configuration.MyConfigurationGood() {
			return &pb.RequestVoteResponse{
				Term:        s.State.GetCurrentTerm(),
				VoteGranted: false,
			}, nil
		}
	}

	if req.Term < s.State.GetCurrentTerm() {
		return &pb.RequestVoteResponse{
			Term:        s.State.GetCurrentTerm(),
			VoteGranted: false,
		}, nil
	}

	if s.State.GetCurrentState() != CANDIDATE {
		return &pb.RequestVoteResponse{
			Term: s.State.GetCurrentTerm(), VoteGranted: false,
		}, nil
	}

	if req.Term > s.State.GetCurrentTerm() {
		s.State.SetCurrentTerm(req.Term)
		s.State.SetCurrentState(FOLLOWER)
	}

	if req.LastLogTerm < s.State.Log.GetMostRecentTerm() {
		return &pb.RequestVoteResponse{
			Term:        s.State.GetCurrentTerm(),
			VoteGranted: false,
		}, nil
	} else if req.LastLogTerm == s.State.Log.GetMostRecentTerm() {
		if req.LastLogIndex < s.State.Log.GetMostRecentIndex() {
			return &pb.RequestVoteResponse{
				Term:        s.State.GetCurrentTerm(),
				VoteGranted: false,
			}, nil
		}
	}

	if s.State.GetVotedFor() == "" || s.State.GetVotedFor() == req.CandidateId {
		s.State.SetVotedFor(req.CandidateId)
		return &pb.RequestVoteResponse{
			Term:        s.State.GetCurrentTerm(),
			VoteGranted: true,
		}, nil
	}

	return &pb.RequestVoteResponse{
		Term:        s.State.GetCurrentTerm(),
		VoteGranted: false,
	}, nil
}

func (s *NetworkServer) getLeader() *Node {
	leaderID := s.State.GetLeaderID()
	if leaderID != "" {
		if s.State.Configuration.InConfiguration(leaderID) {
			node, err := s.State.Configuration.GetNode(leaderID)
			if err == nil {
				return &node
			}
			return nil
		}
	}
	return nil
}

func protoNodesToNodes(protoNodes []*pb.Node) []Node {
	nodes := make([]Node, len(protoNodes))
	for i := 0; i < len(protoNodes); i++ {
		nodes[i] = Node{
			IP:         protoNodes[i].Ip,
			Port:       protoNodes[i].Port,
			CommonName: protoNodes[i].CommonName,
			NodeID:     protoNodes[i].NodeId,
		}
	}
	return nodes
}

func (s *NetworkServer) appendLogEntry(entry *pb.Entry) {
	_, err := s.State.Log.AppendEntry(&pb.LogEntry{
		Term:  s.State.GetCurrentTerm(),
		Entry: entry,
	})
	if err != nil {
		log.Printf("failed to append log entry: %v", err)
		return
	}

	if entry.Type == pb.EntryType_ENTRY_TYPE_CONFIGURATION_CHANGE {
		config := entry.GetConfig()
		if config == nil {
			log.Fatal("Incorrect entry information. No configuration present")
		}
		if config.Type == pb.ConfigurationType_CONFIGURATION_TYPE_CURRENT {
			s.State.Configuration.UpdateCurrentConfiguration(protoNodesToNodes(config.Nodes), s.State.Log.GetMostRecentIndex())
		} else {
			s.State.Configuration.NewFutureConfiguration(protoNodesToNodes(config.Nodes), s.State.Log.GetMostRecentIndex())
		}
	}
}

func (s *NetworkServer) addLogEntryLeader(entry *pb.Entry) error {
	if entry.Type == pb.EntryType_ENTRY_TYPE_CONFIGURATION_CHANGE {
		config := entry.GetConfig()
		if config != nil {
			if config.Type == pb.ConfigurationType_CONFIGURATION_TYPE_FUTURE {
				if s.State.Configuration.GetFutureConfigurationActive() {
					return errors.New("Can not change confirugation while another configuration change is underway")
				}
				if s.Encrypted {
					for _, v := range config.GetNodes() {
						if !keyman.StateMachine.NodeInGeneration(keyman.StateMachine.GetCurrentGeneration(), v.NodeId) {
							return fmt.Errorf("node %s not in current generation (%d)",
								v.NodeId, keyman.StateMachine.GetCurrentGeneration())
						}
					}
				}
			}
		} else {
			return errors.New("Incorrect entry information. No configuration present")
		}
	}
	s.appendLogEntry(entry)
	s.State.calculateNewCommitIndex()
	s.State.SendAppendEntries <- true
	return nil
}

// ClientToLeaderRequest implementation
func (s *NetworkServer) ClientToLeader(ctx context.Context, req *pb.ClientToLeaderRequest) (*pb.ClientToLeaderResponse, error) {
	if !s.State.Configuration.InConfiguration(req.SenderId) {
		return &pb.ClientToLeaderResponse{}, errors.New("Node is not in the configuration")
	}

	if s.State.GetCurrentState() != LEADER {
		return &pb.ClientToLeaderResponse{}, errors.New("Node is not the current leader")
	}
	err := s.addLogEntryLeader(req.Entry)
	return &pb.ClientToLeaderResponse{}, err
}

//sendLeaderLogEntry forwards a client request to the leader
func (s *NetworkServer) sendLeaderLogEntry(entry *pb.Entry) error {
	sendLogTimeout := time.After(leaderRequestTimeout)
	lastError := errors.New("timeout before client to leader request was attempted")
	for {
		select {
		case <-sendLogTimeout:
			return lastError
		default:
			leaderNode := s.getLeader()
			if leaderNode == nil {
				lastError = errors.New("Unable to find leader")
				continue
			}

			conn, err := s.Dial(leaderNode, SendEntryTimeout)
			if err != nil {
				lastError = err
				continue
			}
			defer conn.Close()

			if err == nil {
				client := pb.NewRaftNetworkServiceClient(conn)
				_, err := client.ClientToLeader(context.Background(), &pb.ClientToLeaderRequest{
					SenderId: s.State.NodeID,
					Entry:    entry,
				})
				if err == nil {
					return err
				}
				lastError = err
			}
		}
	}
}

func generateNewUUID() string {
	uuidBytes, err := ioutil.ReadFile("/proc/sys/kernel/random/uuid")
	if err != nil {
		log.Fatalf("Error generating new UUID: %v", err)
	}
	return strings.TrimSpace(string(uuidBytes))
}

func convertNodesToProto(nodes []Node) []*pb.Node {
	protoNodes := make([]*pb.Node, len(nodes))
	for i := 0; i < len(nodes); i++ {
		protoNodes[i] = &pb.Node{
			Ip:         nodes[i].IP,
			Port:       nodes[i].Port,
			CommonName: nodes[i].CommonName,
			NodeId:     nodes[i].NodeID,
		}
	}
	return protoNodes
}

//getRandomElectionTimeout returns a time between ELECTION_TIMEOUT and ELECTION_TIMEOUT*2
func getRandomElectionTimeout() time.Duration {
	rand.Seed(time.Now().UnixNano())
	return ElectionTimeout + time.Duration(rand.Int63n(int64(ElectionTimeout)))
}

func (s *NetworkServer) electionTimeOut() {
	timer := time.NewTimer(getRandomElectionTimeout())
	defer s.Wait.Done()
	defer timer.Stop()
	for {
		select {
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				log.Print("Exiting election timeout loop")
				return
			}
		case <-s.ElectionTimeoutReset:
			timer.Reset(getRandomElectionTimeout())
		case <-timer.C:
			if s.State.Configuration.HasConfiguration() {
				log.Print("Starting new election")
				s.State.SetCurrentTerm(s.State.GetCurrentTerm() + 1)
				s.State.SetCurrentState(CANDIDATE)
				timer.Reset(getRandomElectionTimeout())
			}
		}
	}
}

// Dial a node
func (s *NetworkServer) Dial(node *Node, timeoutMiliseconds time.Duration) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials = insecure.NewCredentials()
	if s.TLSEnabled {
		creds = credentials.NewTLS(&tls.Config{
			ServerName:         s.nodeDetails.CommonName,
			InsecureSkipVerify: s.TLSSkipVerify,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutMiliseconds)
	defer cancel()

	return grpc.DialContext(ctx, node.String(), grpc.WithTransportCredentials(creds))
}

func (s *NetworkServer) requestPeerVote(node *Node, term uint64, voteChannel chan *voteResponse) {
	defer s.Wait.Done()
	for {
		if term != s.State.GetCurrentTerm() || s.State.GetCurrentState() != CANDIDATE {
			voteChannel <- nil
			return
		}
		log.Printf("Dialing %v", node)
		conn, err := s.Dial(node, RequestVoteTimeout)
		if err != nil {
			continue
		}
		defer conn.Close()
		client := pb.NewRaftNetworkServiceClient(conn)
		response, err := client.RequestVote(context.Background(), &pb.RequestVoteRequest{
			Term:         s.State.GetCurrentTerm(),
			CandidateId:  s.State.NodeID,
			LastLogIndex: s.State.Log.GetMostRecentIndex(),
			LastLogTerm:  s.State.Log.GetMostRecentTerm(),
		})
		log.Printf("Got response from %s", node)
		if err == nil {
			voteChannel <- &voteResponse{response, node.NodeID}
			return
		}
	}
}

type voteResponse struct {
	response *pb.RequestVoteResponse
	NodeID   string
}

//runElection attempts to get elected as leader for the current term
func (s *NetworkServer) runElection() {
	defer s.Wait.Done()
	term := s.State.GetCurrentTerm()
	var votesGranted []string

	if s.State.GetVotedFor() == "" && s.State.Configuration.InConfiguration(s.State.NodeID) {
		s.State.SetVotedFor(s.State.NodeID)
		votesGranted = append(votesGranted, s.State.NodeID)
	}

	if s.State.Configuration.HasMajority(votesGranted) {
		log.Printf("Node elected leader with %d votes", len(votesGranted))
		s.State.SetCurrentState(LEADER)
		return
	}

	log.Print("Sending RequestVote RPCs to peers")
	voteChannel := make(chan *voteResponse)
	peers := s.State.Configuration.GetPeersList()
	for i := 0; i < len(peers); i++ {
		s.Wait.Add(1)
		go s.requestPeerVote(&peers[i], term, voteChannel)
	}

	votesReturned := 0
	votesRequested := len(peers)
	if votesRequested == 0 {
		return
	}
	for {
		select {
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				log.Print("Exiting election loop")
				return
			}
		case vote := <-voteChannel:
			if term != s.State.GetCurrentTerm() || s.State.GetCurrentState() != CANDIDATE {
				return
			}
			votesReturned++
			if vote != nil {
				if vote.response.Term > s.State.GetCurrentTerm() {
					log.Print("Stopping election, higher term encountered.")
					s.State.SetCurrentTerm(vote.response.Term)
					s.State.SetCurrentState(FOLLOWER)
					return
				}

				if vote.response.VoteGranted {
					votesGranted = append(votesGranted, vote.NodeID)
					log.Printf("Vote granted. Current votes: %d", len(votesGranted))
					if s.State.Configuration.HasMajority(votesGranted) {
						log.Printf("Node elected leader with %d votes", len(votesGranted))
						s.State.SetCurrentState(LEADER)
						return
					}
				}
			}
			if votesReturned == votesRequested {
				return
			}
		}
	}
}

func (s *NetworkServer) manageElections() {
	defer s.Wait.Done()
	for {
		select {
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				log.Print("Exiting election management loop")
				return
			}
		case <-s.State.StartElection:
			s.Wait.Add(1)
			go s.runElection()
		}
	}
}

//sendHeartBeat is used when we are the leader to both replicate log entries and prevent other nodes from timing out
func (s *NetworkServer) sendHeartBeat(node *Node) {
	defer s.Wait.Done()
	nextIndex := s.State.Configuration.GetNextIndex(node.NodeID)
	sendingSnapshot := s.State.Configuration.GetSendingSnapshot(node.NodeID)

	if s.State.Log.GetStartIndex() >= nextIndex && !sendingSnapshot {
		s.State.SendSnapshot <- *node
		sendingSnapshot = true
	}

	conn, err := s.Dial(node, HeartbeatTimeout)
	if err != nil {
		return
	}
	defer conn.Close()
	client := pb.NewRaftNetworkServiceClient(conn)
	if s.State.Log.GetMostRecentIndex() >= nextIndex && !sendingSnapshot {
		prevLogTerm := uint64(0)
		if nextIndex-1 > 0 {
			prevLogEntry, err := s.State.Log.GetLogEntry(nextIndex - 1)
			if err != nil {
				if err == raftlog.ErrIndexBelowStartIndex {
					prevLogTerm = s.State.Log.GetStartTerm()
				} else {
					log.Fatalf("Unable to get log entry at %d: %v", nextIndex-1, err)
				}
			} else {
				prevLogTerm = prevLogEntry.Term
			}
		}

		nextLogEntries, err := s.State.Log.GetLogEntries(nextIndex, MaxAppendEntries)
		if err != nil {
			if err == raftlog.ErrIndexBelowStartIndex {
				s.State.SendSnapshot <- *node
				return
			}
			log.Fatalf("Unable to get log entry: %v", err)
		}
		numLogEntries := uint64(len(nextLogEntries))

		response, err := client.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
			Term:         s.State.GetCurrentTerm(),
			LeaderId:     s.State.NodeID,
			PrevLogIndex: nextIndex - 1,
			PrevLogTerm:  prevLogTerm,
			Entries:      nextLogEntries,
			LeaderCommit: s.State.GetCommitIndex(),
		})
		if err == nil {
			if response.Term > s.State.GetCurrentTerm() {
				s.State.StopLeading <- true
			} else if !response.Success {
				if s.State.GetCurrentState() == LEADER {
					if response.NextIndex == 0 {
						s.State.Configuration.SetNextIndex(node.NodeID, nextIndex-1)
					} else {
						s.State.Configuration.SetNextIndex(node.NodeID, response.NextIndex)
					}
				}
			} else if response.Success {
				if s.State.GetCurrentState() == LEADER {
					s.State.Configuration.SetNextIndex(node.NodeID, nextIndex+numLogEntries)
					s.State.Configuration.SetMatchIndex(node.NodeID, nextIndex+numLogEntries-1)
					s.State.calculateNewCommitIndex()
				}
			}
		}
	} else {
		response, err := client.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
			Term:         s.State.GetCurrentTerm(),
			LeaderId:     s.State.NodeID,
			PrevLogIndex: s.State.Log.GetMostRecentIndex(),
			PrevLogTerm:  s.State.Log.GetMostRecentTerm(),
			Entries:      []*pb.Entry{},
			LeaderCommit: s.State.GetCommitIndex(),
		})
		if err == nil {
			if response.Term > s.State.GetCurrentTerm() {
				s.State.StopLeading <- true
			} else {
				if !response.Success {
					if s.State.GetCurrentState() == LEADER {
						if response.NextIndex == 0 {
							s.State.Configuration.SetNextIndex(node.NodeID, nextIndex-1)
						} else {
							s.State.Configuration.SetNextIndex(node.NodeID, response.NextIndex)
						}
					}
				}
			}
		}
	}
}

func (s *NetworkServer) manageLeading() {
	defer s.Wait.Done()
	for {
		select {
		//We want to keep these channels clear for when we first become leader
		case <-s.State.StopLeading:
		case <-s.State.SendAppendEntries:
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				s.State.SetCurrentState(INACTIVE)
				log.Print("Exiting leading management loop")
				return
			}
		case <-s.State.StartLeading:
			log.Printf("Started leading for term %d", s.State.GetCurrentTerm())
			s.State.Configuration.ResetNodeIndices(s.State.Log.GetMostRecentIndex())
			peers := s.State.Configuration.GetPeersList()
			for i := 0; i < len(peers); i++ {
				s.Wait.Add(1)
				go s.sendHeartBeat(&peers[i])
			}
			timer := time.NewTimer(Heartbeat)
		leadingLoop:
			for {
				select {
				case _, ok := <-s.Quit:
					if !ok {
						s.QuitChannelClosed = true
						log.Print("Exiting heartbeat loop")
						return
					}
				case <-s.State.StopLeading:
					log.Print("Stopped leading")
					break leadingLoop
				case <-s.State.SendAppendEntries:
					timer.Reset(Heartbeat)
					s.ElectionTimeoutReset <- true
					peers = s.State.Configuration.GetPeersList()
					for i := 0; i < len(peers); i++ {
						s.Wait.Add(1)
						go s.sendHeartBeat(&peers[i])
					}
				case <-timer.C:
					timer.Reset(Heartbeat)
					s.ElectionTimeoutReset <- true
					peers = s.State.Configuration.GetPeersList()
					for i := 0; i < len(peers); i++ {
						s.Wait.Add(1)
						go s.sendHeartBeat(&peers[i])
					}
				}
			}
		}
	}
}

//manageConfigurationChanges performs necessary actions when a configuration has been applied.
//Such as stepping down or creating a new configuration change request
func (s *NetworkServer) manageConfigurationChanges() {
	defer s.Wait.Done()
	for {
		select {
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				log.Print("Exiting configuration management loop")
				return
			}
		case config := <-s.State.ConfigurationApplied:
			log.Printf("New configuration applied: %v", config)
			if config.Type == pb.ConfigurationType_CONFIGURATION_TYPE_CURRENT {
				inConfig := false
				for i := 0; i < len(config.Nodes); i++ {
					if config.Nodes[i].NodeId == s.State.NodeID {
						inConfig = true
						break
					}
				}
				if !inConfig {
					log.Printf("Node not included in current configuration %s", s.State.NodeID)
					s.State.SetCurrentState(FOLLOWER)
				}
			} else {
				if s.State.GetCurrentState() == LEADER {
					newConfig := &pb.Entry{
						Type:    pb.EntryType_ENTRY_TYPE_CONFIGURATION_CHANGE,
						Uuid:    generateNewUUID(),
						Command: nil,
						Config: &pb.Configuration{
							Type:  pb.ConfigurationType_CONFIGURATION_TYPE_CURRENT,
							Nodes: config.Nodes,
						},
					}
					_ = s.addLogEntryLeader(newConfig)
				}
			}
		}
	}
}

func (s *NetworkServer) manageEntryApplication() {
	defer s.Wait.Done()
	for {
		select {
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				log.Print("Exiting entry application management loop")
				return
			}
		case <-s.State.ApplyEntries:
			s.State.ApplyLogEntries()
		}
	}
}

// NewNetworkServer creates a new instance of the raft server
func NewNetworkServer(
	nodeDetails Node,
	pfsDirectory, raftInfoDirectory string,
	testConfiguration *StartConfiguration,
	TLSEnabled, TLSSkipVerify, encrypted bool,
) *NetworkServer {
	raftServer := &NetworkServer{State: newState(nodeDetails, pfsDirectory, raftInfoDirectory, testConfiguration)}
	raftServer.ElectionTimeoutReset = make(chan bool, 100)
	raftServer.Quit = make(chan bool)
	raftServer.QuitChannelClosed = false
	raftServer.nodeDetails = nodeDetails
	raftServer.raftInfoDirectory = raftInfoDirectory
	raftServer.TLSEnabled = TLSEnabled
	raftServer.TLSSkipVerify = TLSSkipVerify
	raftServer.Encrypted = encrypted
	raftServer.ChangeNodeLocation(nodeDetails.NodeID, nodeDetails.IP, nodeDetails.Port)
	raftServer.setupSnapshotDirectory()

	raftServer.Wait.Add(6)
	go raftServer.electionTimeOut()
	go raftServer.manageElections()
	go raftServer.manageLeading()
	go raftServer.manageConfigurationChanges()
	go raftServer.manageSnapshoting()
	go raftServer.manageEntryApplication()
	return raftServer
}
