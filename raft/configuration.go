package raft

import (
	"encoding/json"
	"errors"
	"github.com/cpssd/paranoid/raft/raftlog"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

const (
	PersistentConfigurationFileName string = "persistentConfigFile"
)

//Configuration manages configuration information of a raft server
type Configuration struct {
	myNodeId                  string
	futureConfigurationActive bool

	currentConfiguration []Node
	currentNextIndex     []uint64
	currentMatchIndex    []uint64

	futureConfiguration []Node
	futureNextIndex     []uint64
	futureMatchIndex    []uint64

	raftInfoDirectory    string
	persistentConfigLock sync.Mutex
}

//persistentConfiguration is the configuration information that is saved to disk
type persistentConfiguration struct {
	FutureConfigurationActive bool   `json:"futureconfigactive"`
	CurrentConfiguration      []Node `json:"currentconfig"`
	FutureConfiguration       []Node `json:"futureconfig"`
}

//StartConfiguration is used to start a raft node with a specific congiuration for testing purposes
//or if the node is the first node to join a cluster
type StartConfiguration struct {
	Peers []Node
}

func (c *Configuration) GetNode(nodeID string) (Node, error) {
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

func (c *Configuration) NewFutureConfiguration(nodes []Node, lastLogIndex uint64) {
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

func (c *Configuration) UpdateCurrentConfiguration(nodes []Node, lastLogIndex uint64) {
	if len(nodes) == len(c.futureConfiguration) {
		futureToCurrent := true
		for i := 0; i < len(nodes); i++ {
			if c.inFutureConfiguration(nodes[i].NodeID) == false {
				futureToCurrent = false
				break
			}
		}
		if futureToCurrent {
			c.FutureToCurrentConfiguration()
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

func (c *Configuration) GetFutureConfigurationActive() bool {
	return c.futureConfigurationActive
}

func (c *Configuration) FutureToCurrentConfiguration() {
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
	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID == nodeID {
			return true
		}
	}
	return false
}

func (c *Configuration) inFutureConfiguration(nodeID string) bool {
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID == nodeID {
			return true
		}
	}
	return false
}

func (c *Configuration) InConfiguration(nodeID string) bool {
	return c.inCurrentConfiguration(nodeID) || c.inFutureConfiguration(nodeID)
}

func (c *Configuration) MyConfigurationGood() bool {
	if c.InConfiguration(c.myNodeId) {
		if c.GetTotalPossibleVotes() > 1 {
			return true
		}
	}
	return false
}

func (c *Configuration) GetTotalPossibleVotes() int {
	votes := len(c.currentConfiguration)
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.inCurrentConfiguration(c.futureConfiguration[i].NodeID) == false {
			votes++
		}
	}
	return votes
}

func (c *Configuration) GetPeersList() []Node {
	var peers []Node
	for i := 0; i < len(c.currentConfiguration); i++ {
		if c.currentConfiguration[i].NodeID != c.myNodeId {
			peers = append(peers, c.currentConfiguration[i])
		}
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		if c.futureConfiguration[i].NodeID != c.myNodeId {
			if c.inCurrentConfiguration(c.futureConfiguration[i].NodeID) == false {
				peers = append(peers, c.futureConfiguration[i])
			}
		}
	}
	return peers
}

func (c *Configuration) GetNodesList() []Node {
	peers := c.GetPeersList()
	myNode, err := c.GetNode(c.myNodeId)
	if err == nil {
		return append(peers, myNode)
	}
	return peers
}

func getRequiredVotes(nodeCount int) int {
	return (nodeCount / 2) + 1
}

//Check has a majority of votes have been received given a list of NodeIDs
func (c *Configuration) HasMajority(votesRecieved []string) bool {
	currentRequiredVotes := getRequiredVotes(len(c.currentConfiguration))
	count := 0
	for i := 0; i < len(votesRecieved); i++ {
		if c.inCurrentConfiguration(votesRecieved[i]) {
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
			if c.inFutureConfiguration(votesRecieved[i]) {
				count++
			}
		}
		if count < futureRequiredVotes {
			return false
		}
	}
	return true
}

func (c *Configuration) ResetNodeIndexs(lastLogIndex uint64) {
	for i := 0; i < len(c.currentConfiguration); i++ {
		c.currentNextIndex[i] = lastLogIndex + 1
		c.currentMatchIndex[i] = 0
	}
	for i := 0; i < len(c.futureConfiguration); i++ {
		c.futureNextIndex[i] = lastLogIndex + 1
		c.futureMatchIndex[i] = 0
	}
}

func (c *Configuration) GetNextIndex(nodeID string) uint64 {
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
	Log.Fatal("Could not get NextIndex. Node not found")
	return 0
}

func (c *Configuration) GetMatchIndex(nodeID string) uint64 {
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
	Log.Fatal("Could not get MatchIndex. Node not found")
	return 0
}

func (c *Configuration) SetNextIndex(nodeID string, x uint64) {
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

func (c *Configuration) SetMatchIndex(nodeID string, x uint64) {
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

func (c *Configuration) CalculateNewCommitIndex(lastCommitIndex, term uint64, log *raftlog.RaftLog) uint64 {
	if log.GetMostRecentTerm() != term {
		return lastCommitIndex
	}

	currentMajority := getRequiredVotes(len(c.currentMatchIndex))
	futureMajority := getRequiredVotes(len(c.futureMatchIndex))
	newCommitIndex := lastCommitIndex

	for i := lastCommitIndex + 1; i <= log.GetMostRecentIndex(); i++ {
		logEntry, err := log.GetLogEntry(i)
		if err != nil {
			Log.Fatal("Unable to get log entry:", err)
		}
		if logEntry.Term == term {
			currentCount := 0
			if c.inCurrentConfiguration(c.myNodeId) {
				currentCount = 1
			}
			for j := 0; j < len(c.currentMatchIndex); j++ {
				if c.currentConfiguration[j].NodeID != c.myNodeId {
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
				if c.inFutureConfiguration(c.myNodeId) {
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

	persistentCofigBytes, err := json.Marshal(perState)
	if err != nil {
		Log.Fatal("Error saving persistent state to disk:", err)
	}

	err = ioutil.WriteFile(path.Join(c.raftInfoDirectory, PersistentConfigurationFileName), persistentCofigBytes, 0600)
	if err != nil {
		Log.Fatal("Error saving persistent state to disk:", err)
	}
}

func getPersistentConfiguration(persistentConfigurationFile string) *persistentConfiguration {
	if _, err := os.Stat(persistentConfigurationFile); os.IsNotExist(err) {
		return nil
	}
	persistentFileContents, err := ioutil.ReadFile(persistentConfigurationFile)
	if err != nil {
		Log.Fatal("Error reading persistent state from disk:", err)
	}

	perConfig := &persistentConfiguration{}
	err = json.Unmarshal(persistentFileContents, &perConfig)
	if err != nil {
		Log.Fatal("Error reading persistent state from disk:", err)
	}
	return perConfig
}

func newConfiguration(raftInfoDirectory string, testConfiguration *StartConfiguration, myNodeDetails Node) *Configuration {
	var config *Configuration
	if testConfiguration != nil {
		config = &Configuration{
			myNodeId:          myNodeDetails.NodeID,
			raftInfoDirectory: raftInfoDirectory,

			currentConfiguration: append(testConfiguration.Peers, myNodeDetails),
			currentNextIndex:     make([]uint64, len(testConfiguration.Peers)+1),
			currentMatchIndex:    make([]uint64, len(testConfiguration.Peers)+1),
		}
	} else {
		persistentConfig := getPersistentConfiguration(path.Join(raftInfoDirectory, PersistentConfigurationFileName))
		if persistentConfig != nil {
			config = &Configuration{
				myNodeId:          myNodeDetails.NodeID,
				raftInfoDirectory: raftInfoDirectory,

				currentConfiguration: persistentConfig.CurrentConfiguration,
				currentNextIndex:     make([]uint64, len(persistentConfig.CurrentConfiguration)),
				currentMatchIndex:    make([]uint64, len(persistentConfig.CurrentConfiguration)),

				futureConfigurationActive: persistentConfig.FutureConfigurationActive,
				futureConfiguration:       persistentConfig.FutureConfiguration,
				futureNextIndex:           make([]uint64, len(persistentConfig.FutureConfiguration)),
				futureMatchIndex:          make([]uint64, len(persistentConfig.FutureConfiguration)),
			}
		} else {
			config = &Configuration{
				myNodeId:          myNodeDetails.NodeID,
				raftInfoDirectory: raftInfoDirectory,

				currentConfiguration: []Node{},
				currentNextIndex:     []uint64{},
				currentMatchIndex:    []uint64{},
			}
		}
	}
	return config
}
