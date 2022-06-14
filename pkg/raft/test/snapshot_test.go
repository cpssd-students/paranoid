package test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	"github.com/cpssd-students/paranoid/pkg/libpfs"
	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
	"github.com/cpssd-students/paranoid/pkg/raft"
	"github.com/cpssd-students/paranoid/pkg/raft/rafttestutil"
)

func TestSnapshoting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short testing mode")
	}
	t.Parallel()

	t.Log("Testing snapshoting")
	lis, node1Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(lis)
	node := rafttestutil.SetUpNode("node", "localhost", node1Port, "_")
	t.Log("Listeners set up")

	raftDirectory := rafttestutil.CreateRaftDirectory(path.Join(os.TempDir(), "snapshottest", "node"))
	pfsDirectory := path.Join(os.TempDir(), "snapshottestpfs")
	_ = os.RemoveAll(pfsDirectory)

	if err := os.Mkdir(pfsDirectory, 0700); err != nil {
		t.Fatal("Unable to make pfsdirectory:", err)
	}
	keyman.StateMachine = keyman.NewKSM(pfsDirectory)
	defer func() {
		if err := os.RemoveAll(pfsDirectory); err != nil {
			t.Fatal("Error removing pfsdirectory:", err)
		}
	}()

	code, err := libpfs.InitCommand(pfsDirectory)
	if code != returncodes.OK {
		t.Fatal("Unable to init pfsdirectroy:", err)
	}

	var raftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(raftDirectory, raftServer)
	raftServer, srv := raft.StartRaft(lis, node, pfsDirectory, raftDirectory, &raft.StartConfiguration{Peers: []raft.Node{}})
	defer srv.Stop()
	defer rafttestutil.StopRaftServer(raftServer)

	code, err = raftServer.RequestCreatCommand("test.txt", 0700)
	if code != returncodes.OK {
		t.Fatal("Error performing create command:", err)
	}

	code, _, err = raftServer.RequestWriteCommand("test.txt", 0, 5, []byte("hello"))
	if code != returncodes.OK {
		t.Fatal("Error performing write command:", err)
	}

	t.Log("Taking first snapshot")
	err = raftServer.CreateSnapshot(raftServer.State.Log.GetMostRecentIndex())
	if err != nil {
		t.Fatal("Error taking snapshot:", err)
	}

	code, _, err = raftServer.RequestWriteCommand("test.txt", 0, 7, []byte("goodbye"))
	if code != returncodes.OK {
		t.Fatal("Error performing write command:", err)
	}

	for i := 0; i < 5; i++ {
		_, err = os.Stat(path.Join(raftDirectory, raft.SnapshotDirectory, raft.CurrentSnapshotDirectory))
		if err == nil {
			break
		}
		//Sleep to give time for the snapshot management goroutine to update the current snapshot
		time.Sleep(1 * time.Millisecond)
	}

	t.Log("Reverting to snapshot")
	err = raftServer.RevertToSnapshot(path.Join(raftDirectory, raft.SnapshotDirectory, raft.CurrentSnapshotDirectory))
	if err != nil {
		t.Fatal("Error reverting to snapshot:", err)
	}

	if _, data, _ := libpfs.ReadCommand(pfsDirectory, "test.txt", -1, -1); string(data) != "hello" {
		t.Fatal("Error reverting snapshot. Read does not match 'hello'. Actual:", string(data))
	}

	code, err = raftServer.RequestCreatCommand("test2.txt", 0700)
	if code != returncodes.OK {
		t.Fatal("Error performing create command:", err)
	}

	code, _, err = raftServer.RequestWriteCommand("test2.txt", 0, 5, []byte("world"))
	if code != returncodes.OK {
		t.Fatal("Error performing write command:", err)
	}

	t.Log("Taking second snapshot")
	err = raftServer.CreateSnapshot(raftServer.State.Log.GetMostRecentIndex())
	if err != nil {
		t.Fatal("Error taking snapshot:", err)
	}

	code, _, err = raftServer.RequestWriteCommand("test2.txt", 0, 5, []byte("earth"))
	if code != returncodes.OK {
		t.Fatal("Error performing write command:", err)
	}

	for i := 0; i < 5; i++ {
		_, err = os.Stat(path.Join(raftDirectory, raft.SnapshotDirectory, raft.CurrentSnapshotDirectory))
		if err == nil {
			break
		}
		//Sleep to give time for the snapshot management goroutine to update the current snapshot
		time.Sleep(1 * time.Millisecond)
	}

	t.Log("Reverting to snapshot")
	err = raftServer.RevertToSnapshot(path.Join(raftDirectory, raft.SnapshotDirectory, raft.CurrentSnapshotDirectory))
	if err != nil {
		t.Fatal("Error reverting to snapshot:", err)
	}

	if _, data, _ := libpfs.ReadCommand(pfsDirectory, "test.txt", -1, -1); string(data) != "hello" {
		t.Fatal("Error reverting snapshot. Read does not match 'hello'. Actual:", string(data))
	}

	if _, data, _ := libpfs.ReadCommand(pfsDirectory, "test2.txt", -1, -1); string(data) != "world" {
		t.Fatal("Error reverting snapshot. Read does not match 'world'. Actual:", string(data))
	}
}
