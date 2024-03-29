package raft

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"

	"github.com/cpssd-students/paranoid/pkg/raft/raftlog"
)

// Different constants
const (
	PersistentConfigurationFileName = "persistentConfigFile"
	OriginalConfigurationFileName   = "originalConfigFile"
)

// Configuration manages configuration information of a raft server
type Configuration struct {
	myNodeID                  string
	futureConfigurationActive bool

	currentConfiguration []Node
	currentNextIndex     []uint64
	currentMatchIndex    []uint64

	futureConfiguration []Node
	futureNextIndex     []uint64
	futureMatchIndex    []uint64

	sendingSnapshot map[string]bool

	raftInfoDirectory    string
	persistentConfigLock sync.Mutex
	configLock           sync.Mutex
}

// persistentConfiguration stores the configuration information which is written
// to the disk
type persistentConfiguration struct {
	FutureConfigurationActive bool   `json:"futureconfigactive"`
	CurrentConfiguration      []Node `json:"currentconfig"`
	FutureConfiguration       []Node `json:"futureconfig"`
}

// StartConfiguration is used to start a raft node with a specific congiuration
// for testing purposes or if the node is the first node to join a cluster
type StartConfiguration struct {
	Peers []Node
}

// GetNode based on the node ID. If the node is not found, an error is returned
func (c *Configuration) GetNode(nodeID string) (Node, error) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			return c.currentConfiguration[i], nil
		}
	}

	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			return c.futureConfiguration[i], nil
		}
	}

	return Node{}, errors.New("Node not found in configuration")
}

// ChangeNodeLocation changes the IP and Port of a given nodeID
func (c *Configuration) ChangeNodeLocation(nodeID, IP, Port string) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			c.currentConfiguration[i].IP = IP
			c.currentConfiguration[i].Port = Port
		}
	}

	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			c.futureConfiguration[i].IP = IP
			c.futureConfiguration[i].Port = Port
		}
	}
}

// NewFutureConfiguration creates a future configuration and sets the next index
// of those nodes to lastLogIndex + 1
func (c *Configuration) NewFutureConfiguration(nodes []Node, lastLogIndex uint64) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	c.futureConfigurationActive = true
	c.futureConfiguration = nodes
	c.futureNextIndex = make([]uint64, len(nodes))
	c.futureMatchIndex = make([]uint64, len(nodes))

	c.savePersistentConfiguration()

	for i := 0; i < len(nodes); i++ {
		c.futureNextIndex[i] = lastLogIndex + 1
		c.futureMatchIndex[i] = 0
	}
}

// UpdateCurrentConfiguration updates the current configuration given a set of
// nodes. If all the nodes are in the future configuration, the future
// configuration is changed to the current configuration.
func (c *Configuration) UpdateCurrentConfiguration(nodes []Node, lastLogIndex uint64) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	if len(nodes) == len(c.futureConfiguration) {
		futureToCurrent := true
		for i := 0; i < len(nodes); i++ {
			if !c.inFutureConfigurationUnsafe(nodes[i].NodeID) {
				futureToCurrent = false
				break
			}
		}
		if futureToCurrent {
			c.futureToCurrentConfiguration()
			return
		}
	}

	c.currentConfiguration = nodes
	c.currentNextIndex = make([]uint64, len(nodes))
	c.currentMatchIndex = make([]uint64, len(nodes))
	c.savePersistentConfiguration()

	for i := 0; i < len(nodes); i++ {
		c.currentNextIndex[i] = lastLogIndex + 1
		c.currentMatchIndex[i] = 0
	}
}

// UpdateFromConfigurationFile updates the configuration based on the provided
// file path and the last log index. If the configuration cannot be updated,
// an error is returned
func (c *Configuration) UpdateFromConfigurationFile(
	configurationFilePath string, lastLogIndex uint64,
) error {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	configuration := getPersistentConfiguration(configurationFilePath)
	if configuration == nil {
		return errors.New("unable to update configuration from file")
	}

	c.currentConfiguration = configuration.CurrentConfiguration
	c.currentNextIndex = make([]uint64, len(c.currentConfiguration))
	c.currentMatchIndex = make([]uint64, len(c.currentConfiguration))

	c.futureConfiguration = configuration.FutureConfiguration
	c.futureNextIndex = make([]uint64, len(c.futureConfiguration))
	c.futureMatchIndex = make([]uint64, len(c.futureConfiguration))
	c.futureConfigurationActive = configuration.FutureConfigurationActive

	for i := 0; i < len(c.currentConfiguration); i++ {
		c.currentNextIndex[i] = lastLogIndex + 1
		c.currentMatchIndex[i] = 0
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		c.futureNextIndex[i] = lastLogIndex + 1
		c.futureMatchIndex[i] = 0
	}

	c.savePersistentConfiguration()
	return nil
}

// GetFutureConfigurationActive returns true if the future configuration is
// active
func (c *Configuration) GetFutureConfigurationActive() bool {
	return c.futureConfigurationActive
}

//  futureToCurrentConfiguration changes the current configuration to the future
// configuration and clears the future configuration.
func (c *Configuration) futureToCurrentConfiguration() {
	c.futureConfigurationActive = false
	c.currentConfiguration = c.futureConfiguration
	c.currentNextIndex = c.futureNextIndex
	c.currentMatchIndex = c.futureMatchIndex

	c.futureConfiguration = []Node{}
	c.futureNextIndex = []uint64{}
	c.futureMatchIndex = []uint64{}

	c.savePersistentConfiguration()
}

func (c *Configuration) inCurrentConfiguration(nodeID string) bool {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			return true
		}
	}
	return false
}

// inCurrentConfigurationUnsafe must only be called if the configLock has
// already been locked
func (c *Configuration) inCurrentConfigurationUnsafe(nodeID string) bool {
	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			return true
		}
	}
	return false
}

func (c *Configuration) inFutureConfiguration(nodeID string) bool {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			return true
		}
	}
	return false
}

// inFutureConfigurationUnsafe must only be called if the configLock has already
// been locked
func (c *Configuration) inFutureConfigurationUnsafe(nodeID string) bool {
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			return true
		}
	}
	return false
}

// InConfiguration checks is the node ID in the current or future configuration
func (c *Configuration) InConfiguration(nodeID string) bool {
	return c.inCurrentConfiguration(nodeID) || c.inFutureConfiguration(nodeID)
}

//MyConfigurationGood checks if the configuration contains the current node and
// has more than one member.
func (c *Configuration) MyConfigurationGood() bool {
	if c.InConfiguration(c.myNodeID) {
		if c.GetTotalPossibleVotes() > 1 {
			return true
		}
	}
	return false
}

// HasConfiguration checks if the local node is in the configuration
func (c *Configuration) HasConfiguration() bool {
	return c.InConfiguration(c.myNodeID)
}

// GetTotalPossibleVotes returns the number of nodes in the current
// configuration plus the number in the future configuration not also in the
// current configuration.
func (c *Configuration) GetTotalPossibleVotes() int {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	votes := len(c.currentConfiguration)
	for i := 0; i < len(c.futureConfiguration); i++ {
		if !c.inCurrentConfigurationUnsafe(c.futureConfiguration[i].NodeID) {
			votes++
		}
	}
	return votes
}

// GetPeersList returns a list of all the nodes that must be queried to decide
// on state changes or leader election.
func (c *Configuration) GetPeersList() []Node {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	var peers []Node
	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID != c.myNodeID {
			peers = append(peers, c.currentConfiguration[i])
		}
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID != c.myNodeID {
			if !c.inCurrentConfigurationUnsafe(c.futureConfiguration[i].NodeID) {
				peers = append(peers, c.futureConfiguration[i])
			}
		}
	}
	return peers
}

// GetNodesList returns a list of all the nodes in the cluster including the
// current nodes information.
func (c *Configuration) GetNodesList() []Node {
	peers := c.GetPeersList()
	myNode, err := c.GetNode(c.myNodeID)
	if err == nil {
		return append(peers, myNode)
	}
	return peers
}

func getRequiredVotes(nodeCount int) int {
	return (nodeCount / 2) + 1
}

// HasMajority checks if the majority of votes have been received given a list
// of NodeIDs. A majority is needed in both the current and future
// configurations
func (c *Configuration) HasMajority(votesRecieved []string) bool {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	currentRequiredVotes := getRequiredVotes(len(c.currentConfiguration))
	count := 0
	for i := 0; i < len(votesRecieved); i++ {
		if c.inCurrentConfigurationUnsafe(votesRecieved[i]) {
			count++
		}
	}
	if count < currentRequiredVotes {
		return false
	}

	if c.futureConfigurationActive {
		futureRequiredVotes := getRequiredVotes(len(c.futureConfiguration))
		count = 0
		for i := 0; i < len(votesRecieved); i++ {
			if c.inFutureConfigurationUnsafe(votesRecieved[i]) {
				count++
			}
		}
		if count < futureRequiredVotes {
			return false
		}
	}
	return true
}

// ResetNodeIndices is used to reset the currentIndex and matchindex of each
// peer when elected as a leader.
func (c *Configuration) ResetNodeIndices(lastLogIndex uint64) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		c.currentNextIndex[i] = lastLogIndex + 1
		c.currentMatchIndex[i] = 0
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		c.futureNextIndex[i] = lastLogIndex + 1
		c.futureMatchIndex[i] = 0
	}
}

// GetNextIndex gets the index of the configuration that he node is in.
func (c *Configuration) GetNextIndex(nodeID string) uint64 {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			return c.currentNextIndex[i]
		}
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			return c.futureNextIndex[i]
		}
	}
	log.Fatal("Could not get NextIndex. Node not found")
	return 0
}

// GetMatchIndex returns the index matched by the node.
func (c *Configuration) GetMatchIndex(nodeID string) uint64 {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			return c.currentMatchIndex[i]
		}
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			return c.futureMatchIndex[i]
		}
	}
	log.Fatal("Could not get MatchIndex. Node not found")
	return 0
}

// SetNextIndex sets the next index for the node.
func (c *Configuration) SetNextIndex(nodeID string, x uint64) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			c.currentNextIndex[i] = x
		}
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			c.futureNextIndex[i] = x
		}
	}
}

// SetMatchIndex sets the matched index for the node.
func (c *Configuration) SetMatchIndex(nodeID string, x uint64) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			c.currentMatchIndex[i] = x
		}
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			c.futureMatchIndex[i] = x
		}
	}
}

// GetSendingSnapshot gets the snapshot of the node sending it
// TODO: Verify this
func (c *Configuration) GetSendingSnapshot(nodeID string) bool {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	return c.sendingSnapshot[nodeID]
}

// SetSendingSnapshot of the node
func (c *Configuration) SetSendingSnapshot(nodeID string, x bool) {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	c.sendingSnapshot[nodeID] = x
}

//CalculateNewCommitIndex calculates a new commit index in the manner described
// in the Raft paper
func (c *Configuration) CalculateNewCommitIndex(
	lastCommitIndex, term uint64, rlog *raftlog.RaftLog,
) uint64 {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	if rlog.GetMostRecentTerm() != term {
		return lastCommitIndex
	}

	currentMajority := getRequiredVotes(len(c.currentMatchIndex))
	futureMajority := getRequiredVotes(len(c.futureMatchIndex))
	newCommitIndex := lastCommitIndex

	for i := lastCommitIndex + 1; i <= rlog.GetMostRecentIndex(); i++ {
		logEntry, err := rlog.GetLogEntry(i)
		if err != nil {
			log.Fatal("Unable to get log entry:", err)
		}
		if logEntry.Term == term {
			currentCount := 0
			if c.inCurrentConfigurationUnsafe(c.myNodeID) {
				currentCount = 1
			}
			for j := 0; j < len(c.currentMatchIndex); j++ {
				if c.currentConfiguration[j].NodeID != c.myNodeID {
					if c.currentMatchIndex[j] >= i {
						currentCount++
					}
				}
			}
			if currentCount < currentMajority {
				return newCommitIndex
			}

			if c.futureConfigurationActive {
				futureCount := 0
				if c.inFutureConfigurationUnsafe(c.myNodeID) {
					futureCount = 1
				}
				for j := 0; j < len(c.futureMatchIndex); j++ {
					if c.futureMatchIndex[j] >= i {
						futureCount++
					}
				}
				if futureCount < futureMajority {
					return newCommitIndex
				}

			}
			newCommitIndex = i
		}
	}
	return newCommitIndex
}

func (c *Configuration) savePersistentConfiguration() {
	c.persistentConfigLock.Lock()
	defer c.persistentConfigLock.Unlock()
	perState := &persistentConfiguration{
		FutureConfigurationActive: c.futureConfigurationActive,
		CurrentConfiguration:      c.currentConfiguration,
		FutureConfiguration:       c.futureConfiguration,
	}

	persistentConfigBytes, err := json.Marshal(perState)
	if err != nil {
		log.Fatalf("Error saving persistent confiuration to disk: %v", err)
	}

	if _, err := os.Stat(c.raftInfoDirectory); os.IsNotExist(err) {
		log.Fatalf("Raft Info Directory does not exist: %v", err)
	}

	newPeristentFile := path.Join(c.raftInfoDirectory, PersistentConfigurationFileName+"-new")
	err = ioutil.WriteFile(newPeristentFile, persistentConfigBytes, 0600)
	if err != nil {
		log.Fatalf("Error writing new persistent configuration to disk: %v", err)
	}

	if err = os.Rename(
		newPeristentFile,
		path.Join(c.raftInfoDirectory, PersistentConfigurationFileName),
	); err != nil {
		log.Fatalf("Error saving persistent configuration to disk: %v", err)
	}
}

func (c *Configuration) saveOriginalConfiguration() {
	if err := os.Rename(
		path.Join(c.raftInfoDirectory, PersistentConfigurationFileName),
		path.Join(c.raftInfoDirectory, OriginalConfigurationFileName),
	); err != nil {
		log.Fatalf("Error saving original configuration to disk: %v", err)
	}
	c.savePersistentConfiguration()
}

func getPersistentConfiguration(persistentConfigurationFile string) *persistentConfiguration {
	if _, err := os.Stat(persistentConfigurationFile); os.IsNotExist(err) {
		return nil
	}
	persistentFileContents, err := ioutil.ReadFile(persistentConfigurationFile)
	if err != nil {
		log.Fatalf("Error reading persistent state from disk: %v", err)
	}

	perConfig := &persistentConfiguration{}
	err = json.Unmarshal(persistentFileContents, &perConfig)
	if err != nil {
		log.Fatalf("Error reading persistent state from disk: %v", err)
	}
	return perConfig
}

func newConfiguration(
	raftInfoDirectory string,
	testConfiguration *StartConfiguration,
	myNodeDetails Node,
	saveOriginalConfiguration bool,
) *Configuration {
	var config *Configuration
	loadedPersistentState := false
	if testConfiguration != nil {
		config = &Configuration{
			myNodeID:          myNodeDetails.NodeID,
			raftInfoDirectory: raftInfoDirectory,

			currentConfiguration: append(testConfiguration.Peers, myNodeDetails),
			currentNextIndex:     make([]uint64, len(testConfiguration.Peers)+1),
			currentMatchIndex:    make([]uint64, len(testConfiguration.Peers)+1),
		}
	} else {
		persistentConfig := getPersistentConfiguration(
			path.Join(raftInfoDirectory, PersistentConfigurationFileName))
		if persistentConfig != nil {
			loadedPersistentState = true
			config = &Configuration{
				myNodeID:          myNodeDetails.NodeID,
				raftInfoDirectory: raftInfoDirectory,

				currentConfiguration: persistentConfig.CurrentConfiguration,
				currentNextIndex:     make([]uint64, len(persistentConfig.CurrentConfiguration)),
				currentMatchIndex:    make([]uint64, len(persistentConfig.CurrentConfiguration)),

				futureConfigurationActive: persistentConfig.FutureConfigurationActive,
				futureConfiguration:       persistentConfig.FutureConfiguration,

				futureNextIndex:  make([]uint64, len(persistentConfig.FutureConfiguration)),
				futureMatchIndex: make([]uint64, len(persistentConfig.FutureConfiguration)),
			}
		} else {
			config = &Configuration{
				myNodeID:          myNodeDetails.NodeID,
				raftInfoDirectory: raftInfoDirectory,

				currentConfiguration: []Node{},
				currentNextIndex:     []uint64{},
				currentMatchIndex:    []uint64{},
			}
		}
	}
	config.sendingSnapshot = make(map[string]bool)
	config.savePersistentConfiguration()
	if saveOriginalConfiguration && !loadedPersistentState {
		config.saveOriginalConfiguration()
	}
	return config
}
