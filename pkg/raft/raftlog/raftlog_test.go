package raftlog

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	pb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

var (
	testDir string
)

func createTestDir() {
	err := os.RemoveAll(testDir)
	if err != nil {
		log.Fatalf("error creating test directory: %v", err)
	}

	err = os.Mkdir(testDir, 0777)
	if err != nil {
		log.Fatalf("error creating test directory: %v", err)
	}

	err = os.Mkdir(path.Join(testDir, "meta"), 0777)
	if err != nil {
		log.Fatalf("error creating test directory: %v", err)
	}
}

func removeTestDir() {
	os.RemoveAll(testDir)
}

func TestWriteReadDelete(t *testing.T) {
	testDir = path.Join(os.TempDir(), "paranoidTest")
	removeTestDir()
	createTestDir()

	logDir := path.Join(testDir, "meta", "activity_logs")
	rl := New(logDir)

	// Testing Write functionality
	i, err := rl.AppendEntry(
		&pb.LogEntry{
			Term: 0,
			Entry: &pb.Entry{
				Type: pb.EntryType_ENTRY_TYPE_STATE_MACHINE_COMMAND,
				Command: &pb.StateMachineCommand{
					Type: 0,
					Path: "ThisIsAPath",
				},
			},
		})
	if err != nil || i != 1 {
		t.Error("received error writing log, err:", err)
	}

	i, err = rl.AppendEntry(
		&pb.LogEntry{
			Term: 0,
			Entry: &pb.Entry{
				Type: pb.EntryType_ENTRY_TYPE_STATE_MACHINE_COMMAND,
				Command: &pb.StateMachineCommand{
					Type: 0,
					Path: "ThisIsAPath2",
				},
			},
		})
	if err != nil || i != 2 {
		t.Error("received error writing log, err:", err)
	}

	files, err := ioutil.ReadDir(path.Join(logDir, LogEntryDirectoryName))
	if err != nil {
		t.Error("Error reading log directory")
	}

	if len(files) != 2 {
		t.Error("number of files in directory is not what it should be, writing error.")
	}
	if files[0].Name() != "1000001" || files[1].Name() != "1000002" {
		t.Error("Files not named what they should be, file1: ", files[0].Name(), "file2: ", files[1].Name())
	}

	// Testing Read functionality
	li := rl.GetMostRecentIndex()
	if li != 2 {
		t.Error("Most recent index not what it should be: ", li)
	}

	e, err := rl.GetLogEntry(rl.GetMostRecentIndex())
	if err != nil {
		t.Error("Error received when reading log: ", err)
	}

	if e.GetEntry().GetCommand().Type != 0 || e.GetEntry().GetCommand().Path != "ThisIsAPath2" {
		t.Error("Bad protobuf received from read: ", e)
	}

	entries, _ := rl.GetEntriesSince(1)
	if len(entries) != 2 {
		t.Fatal("Incorrect entries returned")
	}

	if entries[0].GetCommand().Path != "ThisIsAPath" {
		t.Error("Bad protobuf received from getEntiresSince: ", entries[0])
	}

	if entries[1].GetCommand().Path != "ThisIsAPath2" {
		t.Error("Bad protobuf received from getEntiresSince: ", entries[1])
	}

	// Testing Delete functionality
	err = rl.DiscardLogEntriesAfter(1)
	if err != nil {
		t.Error("Error received when deleting log: ", err)
	}

	files, err = ioutil.ReadDir(path.Join(logDir, LogEntryDirectoryName))
	if err != nil {
		t.Error("Error reading log directory")
	}

	if len(files) != 0 {
		t.Error("number of files in directory is not what it should be, delete error.")
	}
}
