// +build !integration

package keystatetest

import (
	"os"
	"path"
	"reflect"
	"testing"

	"paranoid/cmd/pfsd/keyman"
	"paranoid/pkg/libpfs"
	"paranoid/pkg/logger"
	pb "paranoid/pkg/proto/raft"
	"paranoid/pkg/raft"
	"paranoid/pkg/raft/raftlog"
	"paranoid/pkg/raft/rafttestutil"
)

func TestMain(m *testing.M) {
	raft.Log = logger.New("rafttest", "rafttest", os.DevNull)
	raftlog.Log = logger.New("rafttest", "rafttest", os.DevNull)
	libpfs.Log = logger.New("rafttest", "rafttest", os.DevNull)
	keyman.Log = logger.New("rafttest", "rafttest", os.DevNull)
	os.MkdirAll(path.Join(os.TempDir(), "keystatetest", "meta"), 0777)
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestKeyStateUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode.")
	}
	t.Parallel()

	raft.Log.Info("Testing key state updates")
	nodeLis, nodePort := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(nodeLis)
	node := rafttestutil.SetUpNode("node", "localhost", nodePort, "_")
	raft.Log.Info("Node setup complete.")

	nodeRaftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "keystatetest", "node"))
	var nodeRaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(nodeRaftDirectory, nodeRaftServer)
	nodeRaftServer, nodesrv := raft.StartRaft(nodeLis, node, "", nodeRaftDirectory, &raft.StartConfiguration{Peers: []raft.Node{}})
	defer nodesrv.Stop()
	defer rafttestutil.StopRaftServer(nodeRaftServer)

	keyman.StateMachine = keyman.NewKSM(path.Join(os.TempDir(), "keystatetest"))

	pbnode := &pb.Node{
		Ip:         "10.0.0.1",
		Port:       "1337",
		CommonName: "test-node",
		NodeId:     "foobar",
	}
	generation, _, err := keyman.StateMachine.NewGeneration(pbnode.NodeId)
	if err != nil {
		t.Error("Failed to initialise new generation:", err)
	}
	if generation != 0 {
		t.Error("Failed to initialise new generation, exptected generation number 0. Got:", generation)
	}
	if keyman.StateMachine.Generations[0] == nil {
		t.Error("Failed to initialise new generation, generation is nil")
	}
	err = nodeRaftServer.RequestKeyStateUpdate(pbnode, pbnode, 0)
	if err != nil {
		t.Error("RequestKeyStateUpdate returned error:", err)
	}
	if len(keyman.StateMachine.Generations[keyman.StateMachine.CurrentGeneration].Elements) != 1 {
		t.Error("KeyStateMachine not updated. Expected no. elements: 1. Actual:",
			len(keyman.StateMachine.Generations[generation].Elements))
	}

	testMachine, err := keyman.NewKSMFromPFSDir(path.Join(os.TempDir(), "keystatetest"))
	if err != nil {
		t.Error("Failed to create new KSM from PFS directory:", err)
	}
	// Delete the Events channel because we don't care about it.
	keyman.StateMachine.Events = nil
	testMachine.Events = nil
	if !reflect.DeepEqual(keyman.StateMachine, testMachine) {
		t.Log("Decoded and encoded KSM's do not match.")
		t.Log("Expected:", keyman.StateMachine)
		t.Log("Got:", testMachine)
		t.Fail()
	}
}
