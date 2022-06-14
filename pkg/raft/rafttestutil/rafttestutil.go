package rafttestutil

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cpssd-students/paranoid/pkg/raft"
)

// GenerateNewUUID creates a new UUID
func GenerateNewUUID() string {
	uuidBytes, err := ioutil.ReadFile("/proc/sys/kernel/random/uuid")
	if err != nil {
		log.Fatalln("Error generating new UUID:", err)
	}
	return strings.TrimSpace(string(uuidBytes))
}

// StartListener starts a new listener on a random port
func StartListener() (*net.Listener, string) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to start listening : %v.\n", err)
	}
	splits := strings.Split(lis.Addr().String(), ":")
	port := splits[len(splits)-1]
	return &lis, port
}

// SetUpNode sets up a new node
func SetUpNode(name, ip, port, commonName string) raft.Node {
	return raft.Node{
		NodeID:     name,
		IP:         ip,
		Port:       port,
		CommonName: commonName,
	}
}

// CloseListener closes the given net listener
func CloseListener(lis *net.Listener) {
	if lis != nil {
		(*lis).Close()
		file, _ := (*lis).(*net.TCPListener).File()
		file.Close()
	}
}

// StopRaftServer stops the given raft server
func StopRaftServer(raftServer *raft.NetworkServer) {
	if raftServer.QuitChannelClosed == false {
		close(raftServer.Quit)
	}
}

// CreateRaftDirectory clears the specified directory
func CreateRaftDirectory(raftDirectory string) string {
	os.RemoveAll(raftDirectory)
	err := os.MkdirAll(raftDirectory, 0700)
	if err != nil {
		log.Fatal("Error creating raft directory:", err)
	}
	return raftDirectory
}

// RemoveRaftDirectory removes the raft directory when the server is finished
func RemoveRaftDirectory(raftDirectory string, raftServer *raft.NetworkServer) {
	if raftServer != nil {
		//Need to wait for server to shut down or the file could be removed while in use
		raftServer.Wait.Wait()
	}
	time.Sleep(time.Second)
	os.RemoveAll(raftDirectory)
}

// IsLeader checks if the local raft server is the raft leader
func IsLeader(server *raft.NetworkServer) bool {
	return server.State.GetCurrentState() == raft.LEADER
}

// GetLeader from a list of servers
func GetLeader(cluster []*raft.NetworkServer) *raft.NetworkServer {
	highestTerm := uint64(0)
	highestIndex := -1
	for i := 0; i < len(cluster); i++ {
		if IsLeader(cluster[i]) {
			currentTerm := cluster[i].State.GetCurrentTerm()
			if currentTerm > highestTerm {
				highestTerm = currentTerm
				highestIndex = i
			}
		}
	}
	if highestIndex >= 0 {
		return cluster[highestIndex]
	}
	return nil
}

// GetLeaderTimeout gets a leader, if the leader is not found it tires again
// after the timeout
func GetLeaderTimeout(cluster []*raft.NetworkServer, timeoutSeconds int) *raft.NetworkServer {
	var leader *raft.NetworkServer
	leader = GetLeader(cluster)
	if leader != nil {
		return leader
	}
	count := 0
	for {
		count++
		if count > timeoutSeconds {
			break
		}
		time.Sleep(1 * time.Second)
		leader = GetLeader(cluster)
		if leader != nil {
			break
		}
	}
	return leader
}
