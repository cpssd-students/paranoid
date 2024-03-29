package test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cpssd-students/paranoid/pkg/raft"
	"github.com/cpssd-students/paranoid/pkg/raft/rafttestutil"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

func TestRaftElection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short testing mode")
	}
	t.Parallel()

	t.Log("Testing leader eleciton")
	node1Lis, node1Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node1Lis)
	node1 := rafttestutil.SetUpNode("node1", "localhost", node1Port, "_")
	node2Lis, node2Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node2Lis)
	node2 := rafttestutil.SetUpNode("node2", "localhost", node2Port, "_")
	node3Lis, node3Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node3Lis)
	node3 := rafttestutil.SetUpNode("node3", "localhost", node3Port, "_")
	t.Log("Listeners set up")

	node1RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest1", "node1"))
	var node1RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node1RaftDirectory, node1RaftServer)
	node1RaftServer, node1srv := raft.StartRaft(node1Lis, node1, "", node1RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node2, node3}})
	defer node1srv.Stop()
	defer rafttestutil.StopRaftServer(node1RaftServer)

	node2RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest1", "node2"))
	var node2RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node2RaftDirectory, node2RaftServer)
	node2RaftServer, node2srv := raft.StartRaft(node2Lis, node2, "", node2RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node1, node3}})
	defer node2srv.Stop()
	defer rafttestutil.StopRaftServer(node2RaftServer)

	node3RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest1", "node3"))
	var node3RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node3RaftDirectory, node3RaftServer)
	node3RaftServer, node3srv := raft.StartRaft(node3Lis, node3, "", node3RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node1, node2}})
	defer node3srv.Stop()
	defer rafttestutil.StopRaftServer(node3RaftServer)

	cluster := []*raft.NetworkServer{node1RaftServer, node2RaftServer, node3RaftServer}

	t.Log("Searching for leader")
	leader := rafttestutil.GetLeaderTimeout(cluster, 25)
	if leader != nil {
		t.Log(leader.State.NodeID, "selected as leader for term", leader.State.GetCurrentTerm())
	} else {
		t.Fatal("Failed to select leader")
	}

	//Shutdown current leader, make sure an election is triggered and another leader is found
	close(leader.Quit)
	if leader.State.NodeID == "node1" {
		node1srv.Stop()
	} else if leader.State.NodeID == "node2" {
		node2srv.Stop()
	} else {
		node3srv.Stop()
	}
	time.Sleep(5 * time.Second)

	count := 0
	for {
		count++
		if count > 5 {
			t.Fatal("Failed to select leader after original leader is shut down")
		}
		time.Sleep(5 * time.Second)
		newLeader := rafttestutil.GetLeader(cluster)
		if newLeader != nil && leader != newLeader {
			t.Log(newLeader.State.NodeID, "selected as leader for term", newLeader.State.GetCurrentTerm())
			break
		}
	}
}

func verifySpecialNumber(raftServer *raft.NetworkServer, x uint64, waitIntervals int) error {
	if raftServer.State.GetSpecialNumber() == x {
		return nil
	}
	for i := 0; i < waitIntervals; i++ {
		time.Sleep(500 * time.Millisecond)
		if raftServer.State.GetSpecialNumber() == x {
			return nil
		}
	}
	return fmt.Errorf(raftServer.State.NodeID, " special number", raftServer.State.GetSpecialNumber(), " is not equal to", x)
}

func TestRaftLogReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	t.Log("Testing log replication")
	node1Lis, node1Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node1Lis)
	node1 := rafttestutil.SetUpNode("node1", "localhost", node1Port, "_")
	node2Lis, node2Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node2Lis)
	node2 := rafttestutil.SetUpNode("node2", "localhost", node2Port, "_")
	node3Lis, node3Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node3Lis)
	node3 := rafttestutil.SetUpNode("node3", "localhost", node3Port, "_")
	t.Log("Listeners set up")

	node1RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest2", "node1"))
	var node1RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node1RaftDirectory, node1RaftServer)
	node1RaftServer, node1srv := raft.StartRaft(node1Lis, node1, "", node1RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node2, node3}})
	defer node1srv.Stop()
	defer rafttestutil.StopRaftServer(node1RaftServer)

	node2RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest2", "node2"))
	var node2RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node2RaftDirectory, node2RaftServer)
	node2RaftServer, node2srv := raft.StartRaft(node2Lis, node2, "", node2RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node1, node3}})
	defer node2srv.Stop()
	defer rafttestutil.StopRaftServer(node2RaftServer)

	node3RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest2", "node3"))
	var node3RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node3RaftDirectory, node3RaftServer)
	node3RaftServer, node3srv := raft.StartRaft(node3Lis, node3, "", node3RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node1, node2}})
	defer node3srv.Stop()
	defer rafttestutil.StopRaftServer(node3RaftServer)

	_, err := node1RaftServer.RequestAddLogEntry(&pb.Entry{
		Type: pb.EntryType_ENTRY_TYPE_DEMO,
		Uuid: rafttestutil.GenerateNewUUID(),
		Demo: &pb.DemoCommand{Number: 10},
	})
	cluster := []*raft.NetworkServer{node1RaftServer, node2RaftServer, node3RaftServer}
	leader := rafttestutil.GetLeader(cluster)

	if err != nil {
		if leader != nil {
			t.Logf("most recent index: %d", node1RaftServer.State.Log.GetMostRecentIndex())
			t.Logf("most recent leader index: %d", leader.State.Log.GetMostRecentIndex())
			t.Logf("commit index: %d", leader.State.GetCommitIndex())
			t.Logf("leader commit: %d", leader.State.GetCommitIndex())
		}
		t.Fatal("Failed to replicate entry,", err)
	}

	if err = verifySpecialNumber(node1RaftServer, 10, 0); err != nil {
		t.Fatal(err)
	}
	if err = verifySpecialNumber(node2RaftServer, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err = verifySpecialNumber(node3RaftServer, 10, 10); err != nil {
		t.Fatal(err)
	}
}

func TestRaftPersistentState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	t.Log("Testing persistent State")
	node1Lis, node1Port := rafttestutil.StartListener()
	node1 := rafttestutil.SetUpNode("node1", "localhost", node1Port, "_")
	defer rafttestutil.CloseListener(node1Lis)

	node1RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest4", "node1"))
	var node1RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node1RaftDirectory, node1RaftServer)
	node1RaftServer, node1srv := raft.StartRaft(node1Lis, node1, "", node1RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{}})
	defer node1srv.Stop()
	defer rafttestutil.StopRaftServer(node1RaftServer)

	_, err := node1RaftServer.RequestAddLogEntry(&pb.Entry{
		Type: pb.EntryType_ENTRY_TYPE_DEMO,
		Uuid: rafttestutil.GenerateNewUUID(),
		Demo: &pb.DemoCommand{Number: 10},
	})
	if err != nil {
		t.Fatal("Test setup failed,", err)
	}

	cluster := []*raft.NetworkServer{node1RaftServer}

	leader := rafttestutil.GetLeaderTimeout(cluster, 1)
	if leader == nil {
		t.Fatal("Test setup failed: Failed to select leader")
	}

	close(node1RaftServer.Quit)
	node1srv.Stop()
	time.Sleep(1 * time.Second)

	currentTerm := node1RaftServer.State.GetCurrentTerm()
	t.Log("Current Term:", currentTerm)
	lastApplied := node1RaftServer.State.GetLastApplied()
	t.Log("Last applied:", lastApplied)
	votedFor := node1RaftServer.State.GetVotedFor()
	t.Log("Voted For:", votedFor)

	node1RebootLis, _ := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node1RebootLis)

	node1RebootRaftServer, node1Rebootsrv := raft.StartRaft(node1RebootLis, node1, "", node1RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{}})
	defer node1Rebootsrv.Stop()
	defer rafttestutil.StopRaftServer(node1RebootRaftServer)

	if node1RebootRaftServer.State.GetCurrentTerm() != currentTerm {
		t.Fatal("Current term not restored after reboot. CurrentTerm:", node1RebootRaftServer.State.GetCurrentTerm())
	}
	if node1RebootRaftServer.State.GetLastApplied() != lastApplied {
		t.Fatal("Last applied not restored after reboot. Last applied:", node1RebootRaftServer.State.GetLastApplied())
	}
	if node1RebootRaftServer.State.GetVotedFor() != votedFor {
		t.Fatal("Voted for not restored after reboot. Last applied:", node1RebootRaftServer.State.GetVotedFor())
	}
}

func TestRaftConfigurationChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	t.Log("Testing joinging and leaving cluster")

	node1Lis, node1Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node1Lis)
	node1 := rafttestutil.SetUpNode("node1", "localhost", node1Port, "_")

	node2Lis, node2Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node2Lis)
	node2 := rafttestutil.SetUpNode("node2", "localhost", node2Port, "_")

	node3Lis, node3Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node3Lis)
	node3 := rafttestutil.SetUpNode("node3", "localhost", node3Port, "_")

	node4Lis, node4Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node4Lis)
	node4 := rafttestutil.SetUpNode("node4", "localhost", node4Port, "_")

	t.Log("Listeners set up")

	node1RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest3", "node1"))
	var node1RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node1RaftDirectory, node1RaftServer)
	node1RaftServer, node1srv := raft.StartRaft(node1Lis, node1, "", node1RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node2, node3}})
	defer node1srv.Stop()
	defer rafttestutil.StopRaftServer(node1RaftServer)

	node2RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest3", "node2"))
	var node2RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node2RaftDirectory, node2RaftServer)
	node2RaftServer, node2srv := raft.StartRaft(node2Lis, node2, "", node2RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node1, node3}})
	defer node2srv.Stop()
	defer rafttestutil.StopRaftServer(node2RaftServer)

	node3RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest3", "node3"))
	var node3RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node3RaftDirectory, node3RaftServer)
	node3RaftServer, node3srv := raft.StartRaft(node3Lis, node3, "", node3RaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{node1, node2}})
	defer node3srv.Stop()
	defer rafttestutil.StopRaftServer(node3RaftServer)

	node4RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest3", "node4"))
	var node4RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node4RaftDirectory, node4RaftServer)
	node4RaftServer, node4srv := raft.StartRaft(node4Lis, node4, "", node4RaftDirectory, nil)
	defer node4srv.Stop()
	defer rafttestutil.StopRaftServer(node4RaftServer)

	_, err := node1RaftServer.RequestAddLogEntry(&pb.Entry{
		Type: pb.EntryType_ENTRY_TYPE_DEMO,
		Uuid: rafttestutil.GenerateNewUUID(),
		Demo: &pb.DemoCommand{Number: 10},
	})
	if err != nil {
		t.Fatal("Test setup failed:", err)
	}

	err = node2RaftServer.RequestChangeConfiguration([]raft.Node{node1, node2, node3, node4})
	if err != nil {
		t.Fatal("Unabled to change configuration:", err)
	}

	err = verifySpecialNumber(node4RaftServer, 10, 10)
	if err != nil {
		t.Fatal(err)
	}

	cluster := []*raft.NetworkServer{node1RaftServer, node2RaftServer, node3RaftServer, node4RaftServer}

	leader := rafttestutil.GetLeaderTimeout(cluster, 15)
	if leader == nil {
		t.Fatal("Unable to find a leader")
	}

	t.Logf("%s is to leave configuration", leader.State.NodeID)

	newNodes := []raft.Node{node1, node2, node3, node4}
	for i := 0; i < len(newNodes); i++ {
		if newNodes[i].NodeID == leader.State.NodeID {
			newNodes = append(newNodes[:i], newNodes[i+1:]...)
			t.Log("Removing leader from new set of nodes")
			break
		}
	}

	err = node1RaftServer.RequestChangeConfiguration(newNodes)
	if err != nil {
		t.Fatal("Unable to change configuration:", err)
	}

	count := 0
	var newLeader *raft.NetworkServer
	for {
		count++
		if count > 10 {
			t.Fatal("Old leader did not stepdown in time")
		}
		time.Sleep(2 * time.Second)
		newLeader = rafttestutil.GetLeader(cluster)
		if newLeader != nil && newLeader.State.NodeID != leader.State.NodeID {
			break
		}
	}

	time.Sleep(raft.HeartbeatTimeout * 2)

	if node1RaftServer.State.NodeID != leader.State.NodeID {
		_, err := node1RaftServer.RequestAddLogEntry(&pb.Entry{
			Type: pb.EntryType_ENTRY_TYPE_DEMO,
			Uuid: rafttestutil.GenerateNewUUID(),
			Demo: &pb.DemoCommand{Number: 1337},
		})
		if err != nil {
			t.Fatal("Unable to commit new entry:", err)
		}
	} else {
		_, err := node2RaftServer.RequestAddLogEntry(&pb.Entry{
			Type: pb.EntryType_ENTRY_TYPE_DEMO,
			Uuid: rafttestutil.GenerateNewUUID(),
			Demo: &pb.DemoCommand{Number: 1337},
		})
		if err != nil {
			t.Fatal("Unable to commit new entry:", err)
		}
	}
}

func TestStartNodeOutOfConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	t.Parallel()

	t.Log("Testing starting a node without a configuration")

	node1Lis, node1Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node1Lis)
	node1 := rafttestutil.SetUpNode("node1", "localhost", node1Port, "_")

	node1RaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "rafttest5", "node1"))
	var node1RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node1RaftDirectory, node1RaftServer)
	node1RaftServer, node1srv := raft.StartRaft(node1Lis, node1, "", node1RaftDirectory, nil)
	defer node1srv.Stop()
	defer rafttestutil.StopRaftServer(node1RaftServer)

	_, err := node1RaftServer.RequestAddLogEntry(&pb.Entry{
		Type: pb.EntryType_ENTRY_TYPE_DEMO,
		Uuid: rafttestutil.GenerateNewUUID(),
		Demo: &pb.DemoCommand{Number: 10},
	})
	if err == nil {
		t.Fatal("Node should not be able to commit entry outside of configuration")
	}
}
