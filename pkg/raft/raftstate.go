package raft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	"github.com/cpssd-students/paranoid/pkg/raft/raftlog"
	pb "github.com/cpssd-students/paranoid/proto/raft"
)

// NodeType provides information about the type of node
type NodeType int

// Different node types
const (
	FOLLOWER NodeType = iota
	CANDIDATE
	LEADER
	INACTIVE
)

// Constants used by State
const (
	PersistentStateFileName string = "persistentStateFile"
	LogDirectory            string = "raft_logs"
)

// Node data
type Node struct {
	IP         string
	Port       string
	CommonName string
	NodeID     string
}

func (n Node) String() string {
	return fmt.Sprintf("%s:%s", n.IP, n.Port)
}

// State information
type State struct {
	//Used for testing purposes
	specialNumber uint64

	NodeID       string
	pfsDirectory string
	currentState NodeType

	currentTerm uint64
	votedFor    string
	Log         *raftlog.RaftLog
	commitIndex uint64
	lastApplied uint64

	leaderID      string
	Configuration *Configuration

	StartElection     chan bool
	StartLeading      chan bool
	StopLeading       chan bool
	SendAppendEntries chan bool
	ApplyEntries      chan bool
	LeaderElected     chan bool

	snapshotCounter       int
	performingSnapshot    bool
	SendSnapshot          chan Node
	NewSnapshotCreated    chan bool
	SnapshotCounterAtZero chan bool

	waitingForApplied    bool
	EntryApplied         chan *EntryAppliedInfo
	ConfigurationApplied chan *pb.Configuration

	raftInfoDirectory   string
	persistentStateLock sync.Mutex
	stateChangeLock     sync.Mutex
	ApplyEntryLock      sync.Mutex
}

// GetCurrentTerm returns the current term
func (s *State) GetCurrentTerm() uint64 {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.currentTerm
}

// SetCurrentTerm sets the current term
func (s *State) SetCurrentTerm(x uint64) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()

	s.votedFor = ""
	s.currentTerm = x
	s.savePersistentState()
}

// GetCurrentState of the State
func (s *State) GetCurrentState() NodeType {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.currentState
}

// SetCurrentState of the State
func (s *State) SetCurrentState(x NodeType) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()

	if s.currentState == LEADER {
		s.StopLeading <- true
	}
	s.currentState = x
	if x == CANDIDATE {
		s.StartElection <- true
	}
	if x == LEADER {
		s.setLeaderIDUnsafe(s.NodeID)
		s.StartLeading <- true
	}
}

// GetPerformingSnapshot from the State
func (s *State) GetPerformingSnapshot() bool {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.performingSnapshot
}

// SetPerformingSnapshot of the State
func (s *State) SetPerformingSnapshot(x bool) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.performingSnapshot = x
}

// IncrementSnapshotCounter updates the counter
func (s *State) IncrementSnapshotCounter() {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.snapshotCounter++
}

// DecrementSnapshotCounter reduces the snapshot counter. If the snapshotCounter
// reaches 0, SnapshotCounterAtZero is notified
func (s *State) DecrementSnapshotCounter() {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.snapshotCounter--
	if s.snapshotCounter == 0 {
		s.SnapshotCounterAtZero <- true
	}
}

// GetSnapshotCounterValue returns the current snapshot counter
func (s *State) GetSnapshotCounterValue() int {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.snapshotCounter
}

// GetCommitIndex returns the current commit index
func (s *State) GetCommitIndex() uint64 {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.commitIndex
}

// SetCommitIndex sets the current commit index to a given value
func (s *State) SetCommitIndex(x uint64) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.commitIndex = x
	s.SendAppendEntries <- true
	s.ApplyEntries <- true
}

//setCommitIndexUnsafe must only be used when the stateChangeLock has already been locked
func (s *State) setCommitIndexUnsafe(x uint64) {
	s.commitIndex = x
	s.SendAppendEntries <- true
	s.ApplyEntries <- true
}

// SetWaitingForApplied set the value
func (s *State) SetWaitingForApplied(x bool) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.waitingForApplied = x
}

// GetWaitingForApplied sets the value
func (s *State) GetWaitingForApplied() bool {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.waitingForApplied
}

// GetVotedFor returns the voted for
func (s *State) GetVotedFor() string {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.votedFor
}

// SetVotedFor sets the voted for
func (s *State) SetVotedFor(x string) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.votedFor = x
	s.savePersistentState()
}

// GetLeaderID returns the ID of the leader
func (s *State) GetLeaderID() string {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.leaderID
}

// setLeaderIDUnsafe must only be used when the stateChangeLock has already been
// locked
func (s *State) setLeaderIDUnsafe(x string) {
	if s.leaderID == "" {
		s.LeaderElected <- true
	}
	s.leaderID = x
}

// SetLeaderID sets the ID of the leader
func (s *State) SetLeaderID(x string) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()

	if s.leaderID == "" {
		s.LeaderElected <- true
	}
	s.leaderID = x
}

// GetLastApplied returns the last applied
func (s *State) GetLastApplied() uint64 {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.lastApplied
}

// SetLastApplied sets the last applied and saves the state
func (s *State) SetLastApplied(x uint64) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.lastApplied = x
	s.savePersistentState()
}

// setLastAppliedUnsafe must only be used when the stateChangeLock has already
// been locked
func (s *State) setLastAppliedUnsafe(x uint64) {
	s.lastApplied = x
	s.savePersistentState()
}

// SetSpecialNumber sets the number and saves the state
func (s *State) SetSpecialNumber(x uint64) {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.specialNumber = x
	s.savePersistentState()
}

// GetSpecialNumber from the raft state
func (s *State) GetSpecialNumber() uint64 {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	return s.specialNumber
}

func (s *State) applyLogEntry(logEntry *pb.LogEntry) *StateMachineResult {
	switch logEntry.Entry.Type {
	case pb.Entry_Demo:
		demoCommand := logEntry.Entry.GetDemo()
		if demoCommand == nil {
			Log.Fatal("Error applying Log to state machine")
		}
		s.specialNumber = demoCommand.Number
	case pb.Entry_ConfigurationChange:
		config := logEntry.Entry.GetConfig()
		if config != nil {
			s.ConfigurationApplied <- config
		} else {
			Log.Fatal("Error applying configuration update")
		}
	case pb.Entry_StateMachineCommand:
		libpfsCommand := logEntry.Entry.GetCommand()
		if libpfsCommand == nil {
			Log.Fatal("Error applying Log to state machine")
		}
		if s.pfsDirectory == "" {
			Log.Fatal("PfsDirectory is not set")
		}
		return PerformLibPfsCommand(s.pfsDirectory, libpfsCommand)
	case pb.Entry_KeyStateCommand:
		keyCommand := logEntry.Entry.GetKeyCommand()
		if keyCommand == nil {
			Log.Fatal("Error applying KeyStateCommand to state machine")
		}
		return PerformKSMCommand(keyman.StateMachine, keyCommand)
	}
	return nil
}

//ApplyLogEntries applys all log entries that have been committed but not yet applied
func (s *State) ApplyLogEntries() {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()
	s.ApplyEntryLock.Lock()
	defer s.ApplyEntryLock.Unlock()

	if s.commitIndex > s.lastApplied {
		for i := s.lastApplied + 1; i <= s.commitIndex; i++ {
			LogEntry, err := s.Log.GetLogEntry(i)
			if err != nil {
				Log.Fatal("Unable to get log entry1:", err)
			}
			result := s.applyLogEntry(LogEntry)
			s.setLastAppliedUnsafe(i)
			if s.waitingForApplied {
				s.EntryApplied <- &EntryAppliedInfo{
					Index:  i,
					Result: result,
				}
			}
		}
	}
}

func (s *State) calculateNewCommitIndex() {
	s.stateChangeLock.Lock()
	defer s.stateChangeLock.Unlock()

	newCommitIndex := s.Configuration.CalculateNewCommitIndex(s.commitIndex, s.currentTerm, s.Log)
	if newCommitIndex > s.commitIndex {
		s.setCommitIndexUnsafe(newCommitIndex)
	}
}

type persistentState struct {
	SpecialNumber uint64 `json:"specialnumber"`
	CurrentTerm   uint64 `json:"currentterm"`
	VotedFor      string `json:"votedfor"`
	LastApplied   uint64 `json:"lastapplied"`
}

func (s *State) savePersistentState() {
	s.persistentStateLock.Lock()
	defer s.persistentStateLock.Unlock()

	perState := &persistentState{
		SpecialNumber: s.specialNumber,
		CurrentTerm:   s.currentTerm,
		VotedFor:      s.votedFor,
		LastApplied:   s.lastApplied,
	}

	persistentStateBytes, err := json.Marshal(perState)
	if err != nil {
		Log.Fatal("Error saving persistent state to disk:", err)
	}

	if _, err := os.Stat(s.raftInfoDirectory); os.IsNotExist(err) {
		Log.Fatal("Raft Info Directory does not exist:", err)
	}

	newPeristentFile := path.Join(s.raftInfoDirectory, PersistentStateFileName+"-new")
	err = ioutil.WriteFile(newPeristentFile, persistentStateBytes, 0600)
	if err != nil {
		Log.Fatal("Error writing new persistent state to disk:", err)
	}

	err = os.Rename(newPeristentFile, path.Join(s.raftInfoDirectory, PersistentStateFileName))
	if err != nil {
		Log.Fatal("Error saving persistent state to disk:", err)
	}
}

func getPersistentState(persistentStateFile string) *persistentState {
	if _, err := os.Stat(persistentStateFile); os.IsNotExist(err) {
		return nil
	}
	persistentFileContents, err := ioutil.ReadFile(persistentStateFile)
	if err != nil {
		Log.Fatal("Error reading persistent state from disk:", err)
	}

	perState := &persistentState{}
	err = json.Unmarshal(persistentFileContents, &perState)
	if err != nil {
		Log.Fatal("Error reading persistent state from disk:", err)
	}
	return perState
}

func newState(myNodeDetails Node, pfsDirectory, raftInfoDirectory string, testConfiguration *StartConfiguration) *State {
	persistentState := getPersistentState(path.Join(raftInfoDirectory, PersistentStateFileName))
	var state *State
	if persistentState == nil {
		state = &State{
			specialNumber:      0,
			pfsDirectory:       pfsDirectory,
			NodeID:             myNodeDetails.NodeID,
			currentTerm:        0,
			votedFor:           "",
			Log:                raftlog.New(path.Join(raftInfoDirectory, LogDirectory)),
			commitIndex:        0,
			lastApplied:        0,
			leaderID:           "",
			snapshotCounter:    0,
			performingSnapshot: false,
			Configuration:      newConfiguration(raftInfoDirectory, testConfiguration, myNodeDetails, true),
			raftInfoDirectory:  raftInfoDirectory,
		}
	} else {
		state = &State{
			specialNumber:      persistentState.SpecialNumber,
			pfsDirectory:       pfsDirectory,
			NodeID:             myNodeDetails.NodeID,
			currentTerm:        persistentState.CurrentTerm,
			votedFor:           persistentState.VotedFor,
			Log:                raftlog.New(path.Join(raftInfoDirectory, LogDirectory)),
			commitIndex:        0,
			lastApplied:        persistentState.LastApplied,
			leaderID:           "",
			snapshotCounter:    0,
			performingSnapshot: false,
			Configuration:      newConfiguration(raftInfoDirectory, testConfiguration, myNodeDetails, true),
			raftInfoDirectory:  raftInfoDirectory,
		}
	}

	state.StartElection = make(chan bool, 100)
	state.StartLeading = make(chan bool, 100)
	state.StopLeading = make(chan bool, 100)
	state.SendAppendEntries = make(chan bool, 100)
	state.ApplyEntries = make(chan bool, 100)
	state.LeaderElected = make(chan bool, 1)
	state.EntryApplied = make(chan *EntryAppliedInfo, 100)
	state.NewSnapshotCreated = make(chan bool, 100)
	state.SnapshotCounterAtZero = make(chan bool, 100)
	state.SendSnapshot = make(chan Node, 100)
	state.ConfigurationApplied = make(chan *pb.Configuration, 100)
	return state
}
