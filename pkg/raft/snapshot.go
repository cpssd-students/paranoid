package raft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	"github.com/cpssd-students/paranoid/pkg/libpfs"
	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

// String constants
const (
	SnapshotDirectory        = "snapshots"
	CurrentSnapshotDirectory = "currentsnapshot"
	SnapshotMetaFileName     = "snapshotmeta"
	TarFileName              = "snapshot.tar"
)

// Snapshot constants
const (
	SnapshortInteval        time.Duration = 1 * time.Minute
	SnapshotLogsize         uint64        = 2 * 1024 * 1024 //2 MegaBytes
	SnapshotChunkSize       int64         = 1024
	MaxInstallSnapshotFails int           = 10
)

// Called every time raft network server starts up
// Makes sure snapshot directory exists and we have access to it
func (s *NetworkServer) setupSnapshotDirectory() {
	_, err := os.Stat(path.Join(s.raftInfoDirectory, SnapshotDirectory))
	if os.IsNotExist(err) {
		err := os.Mkdir(path.Join(s.raftInfoDirectory, SnapshotDirectory), 0700)
		if err != nil {
			log.Fatalf("failed to create snapshot directory: %v", err)
		}
	} else if err != nil {
		log.Fatalf("error accessing snapshot directory: %v", err)
	}
}

// SnapShotInfo contains the information about the snapshot
type SnapShotInfo struct {
	LastIncludedIndex uint64 `json:"lastincludedindex"`
	LastIncludedTerm  uint64 `json:"lastincludedterm"`
	SelfCreated       bool   `json:"selfcreated"`
}

func getSnapshotMetaInformation(snapShotPath string) (*SnapShotInfo, error) {
	metaFileContents, err := ioutil.ReadFile(path.Join(snapShotPath, SnapshotMetaFileName))
	snapShotInfo := &SnapShotInfo{}
	if err != nil {
		return snapShotInfo, fmt.Errorf("error reading raft meta information: %w", err)
	}

	err = json.Unmarshal(metaFileContents, &snapShotInfo)
	if err != nil {
		return snapShotInfo, fmt.Errorf("error reading raft meta information: %w", err)
	}
	return snapShotInfo, nil
}

func saveSnapshotMetaInformation(snapShotPath string, lastIncludedIndex, lastIncludedTerm uint64, selfCreated bool) error {
	snapShotInfo := &SnapShotInfo{
		LastIncludedIndex: lastIncludedIndex,
		LastIncludedTerm:  lastIncludedTerm,
		SelfCreated:       selfCreated,
	}

	snapShotInfoJSON, err := json.Marshal(snapShotInfo)
	if err != nil {
		return fmt.Errorf("error saving snapshot meta information: %w", err)
	}

	err = ioutil.WriteFile(path.Join(snapShotPath, SnapshotMetaFileName), snapShotInfoJSON, 0600)
	if err != nil {
		return fmt.Errorf("error saving snapshot meta information: %w", err)
	}

	return nil
}

func unpackTarFile(tarFilePath, directory string) error {
	untar := exec.Command("tar", "-xf", tarFilePath, "--directory="+directory)
	err := untar.Run()
	if err != nil {
		return fmt.Errorf("error unarchiving %s: %w", tarFilePath, err)
	}

	err = os.RemoveAll(path.Join(directory, "contents"))
	if err != nil {
		return fmt.Errorf("error unpacking %s: %w", tarFilePath, err)
	}
	err = os.Rename(path.Join(directory, "contents-tar"), path.Join(directory, "contents"))
	if err != nil {
		return fmt.Errorf("error unpacking %s: %w", tarFilePath, err)
	}

	err = os.RemoveAll(path.Join(directory, "names"))
	if err != nil {
		return fmt.Errorf("error unpacking %s: %w", tarFilePath, err)
	}
	err = os.Rename(path.Join(directory, "names-tar"), path.Join(directory, "names"))
	if err != nil {
		return fmt.Errorf("error unpacking %s: %w", tarFilePath, err)
	}

	err = os.RemoveAll(path.Join(directory, "inodes"))
	if err != nil {
		return fmt.Errorf("error unpacking %s: %w", tarFilePath, err)
	}
	err = os.Rename(path.Join(directory, "inodes-tar"), path.Join(directory, "inodes"))
	if err != nil {
		return fmt.Errorf("error unpacking %s: %w", tarFilePath, err)
	}
	return nil
}

func tarSnapshot(snapshotDirectory string) error {
	err := os.Rename(path.Join(snapshotDirectory, "contents"), path.Join(snapshotDirectory, "contents-tar"))
	if err != nil {
		return fmt.Errorf("error creating tar file: %s", err)
	}

	err = os.Rename(path.Join(snapshotDirectory, "inodes"), path.Join(snapshotDirectory, "inodes-tar"))
	if err != nil {
		return fmt.Errorf("error creating tar file: %s", err)
	}

	err = os.Rename(path.Join(snapshotDirectory, "names"), path.Join(snapshotDirectory, "names-tar"))
	if err != nil {
		return fmt.Errorf("error creating tar file: %s", err)
	}

	err = os.Rename(path.Join(snapshotDirectory, "meta", keyman.KsmFileName),
		path.Join(snapshotDirectory, "meta", keyman.KsmFileName+"-tar"))
	if err != nil {
		return fmt.Errorf("error creating tar file: %s", err)
	}

	tar := exec.Command("tar", "--directory="+snapshotDirectory, "-cf", path.Join(snapshotDirectory, TarFileName),
		"contents-tar", "names-tar", "inodes-tar", PersistentConfigurationFileName, path.Join("meta", keyman.KsmFileName+"-tar"))
	err = tar.Run()
	if err != nil {
		return fmt.Errorf("error creating tar file: %s", err)
	}

	return nil
}

func startNextSnapshotWithCurrent(currentSnapshot, nextSnapshot string) error {
	err := unpackTarFile(path.Join(currentSnapshot, TarFileName), nextSnapshot)
	if err != nil {
		return fmt.Errorf("error starting new snapshot from current snapshot: %v", err)
	}

	err = os.Rename(path.Join(nextSnapshot, "meta", keyman.KsmFileName+"-tar"),
		path.Join(nextSnapshot, "meta", keyman.KsmFileName))
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	_, err = os.Create(path.Join(nextSnapshot, "meta", "lock"))
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	return nil
}

func copyFile(originalPath, copyPath string) error {
	contents, err := ioutil.ReadFile(originalPath)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	err = ioutil.WriteFile(copyPath, contents, 0600)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	return nil
}

func (s *NetworkServer) startNextSnapshot(nextSnapshot string) error {
	err := os.Mkdir(path.Join(nextSnapshot, "contents"), 0700)
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	err = os.Mkdir(path.Join(nextSnapshot, "inodes"), 0700)
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	err = os.Mkdir(path.Join(nextSnapshot, "names"), 0700)
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	err = os.Mkdir(path.Join(nextSnapshot, "meta"), 0700)
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	_, err = os.Create(path.Join(nextSnapshot, "meta", "lock"))
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	err = copyFile(path.Join(s.raftInfoDirectory, OriginalConfigurationFileName), path.Join(nextSnapshot, PersistentConfigurationFileName))
	if err != nil {
		return fmt.Errorf("error starting next snapshot: %s", err)
	}

	return nil
}

func (s *NetworkServer) applyLogUpdates(snapshotDirectory string, startIndex, endIndex uint64) (lastIncludedTerm uint64, err error) {
	snapshotConfig := newConfiguration(snapshotDirectory, nil, s.nodeDetails, false)
	var snapshotKeyMachine *keyman.KeyStateMachine
	_, err = os.Stat(path.Join(snapshotDirectory, "meta", keyman.KsmFileName))
	if os.IsNotExist(err) {
		snapshotKeyMachine = keyman.NewKSM(snapshotDirectory)
		err = snapshotKeyMachine.SerialiseToPFSDir()
		if err != nil {
			return 0, fmt.Errorf("unable to apply log entry: %s", err)
		}
	} else if err != nil {
		return 0, fmt.Errorf("unable to apply log entry: %s", err)
	} else {
		snapshotKeyMachine, err = keyman.NewKSMFromPFSDir(snapshotDirectory)
		if err != nil {
			return 0, fmt.Errorf("unable to apply log entry: %s", err)
		}
	}

	if startIndex > endIndex {
		return 0, errors.New("no log entries to apply")
	}

	for i := startIndex; i <= endIndex; i++ {
		logEntry, err := s.State.Log.GetLogEntry(i)
		if err != nil {
			return 0, fmt.Errorf("unable to apply log entry: %s", err)
		}
		if i == endIndex {
			lastIncludedTerm = logEntry.Term
		}

		switch logEntry.Entry.Type {
		case pb.EntryType_ENTRY_TYPE_STATE_MACHINE_COMMAND:
			libpfsCommand := logEntry.Entry.GetCommand()
			if libpfsCommand == nil {
				return 0, errors.New("unable to apply log entry with empty command field")
			}

			result := PerformLibPfsCommand(snapshotDirectory, libpfsCommand)
			if result.Code == returncodes.EUNEXPECTED {
				return 0, fmt.Errorf("error applying log entry: %s", result.Err)
			}
		case pb.EntryType_ENTRY_TYPE_CONFIGURATION_CHANGE:
			config := logEntry.Entry.GetConfig()
			if config == nil {
				return 0, errors.New("unable to apply log entry with empty config field")
			}

			if config.Type == pb.ConfigurationType_CONFIGURATION_TYPE_CURRENT {
				snapshotConfig.UpdateCurrentConfiguration(protoNodesToNodes(config.Nodes), 0)
			} else {
				snapshotConfig.NewFutureConfiguration(protoNodesToNodes(config.Nodes), 0)
			}
		case pb.EntryType_ENTRY_TYPE_KEY_STATE_COMMAND:
			keyChange := logEntry.Entry.GetKeyCommand()
			if keyChange == nil {
				return 0, errors.New("unable to apply log entry with empty key change field")
			}
			result := PerformKSMCommand(snapshotKeyMachine, keyChange)
			if result.Err != nil {
				return 0, fmt.Errorf("error applying log entry: %s", result.Err)
			}
		default:
			return 0, fmt.Errorf("unable to snapshot command type %s", logEntry.Entry.Type)
		}
	}

	return lastIncludedTerm, nil
}

//performCleanup is used to clean up tempory files used in snapshot creation
func performCleanup(snapshotPath string) error {
	err := os.RemoveAll(path.Join(snapshotPath, "contents-tar"))
	if err != nil {
		return fmt.Errorf("error cleaning up temporary files: %s", err)
	}

	err = os.RemoveAll(path.Join(snapshotPath, "inodes-tar"))
	if err != nil {
		return fmt.Errorf("error cleaning up temporary files: %s", err)
	}

	err = os.RemoveAll(path.Join(snapshotPath, "names-tar"))
	if err != nil {
		return fmt.Errorf("error cleaning up temporary files: %s", err)
	}

	err = os.RemoveAll(path.Join(snapshotPath, "meta"))
	if err != nil {
		return fmt.Errorf("error cleaning up temporary files: %s", err)
	}

	err = os.Remove(path.Join(snapshotPath, PersistentConfigurationFileName))
	if err != nil {
		return fmt.Errorf("error cleaning up temporary files: %s", err)
	}
	return nil
}

// CreateSnapshot creates a new snapshot up to the last included index
func (s *NetworkServer) CreateSnapshot(lastIncludedIndex uint64) (err error) {
	currentSnapshot := path.Join(s.raftInfoDirectory, SnapshotDirectory, CurrentSnapshotDirectory)
	nextSnapshot := path.Join(s.raftInfoDirectory, SnapshotDirectory, generateNewUUID())

	if s.State.GetPerformingSnapshot() {
		return errors.New("snapshot creation already in progress")
	}
	s.State.SetPerformingSnapshot(true)
	defer s.State.SetPerformingSnapshot(false)

	err = os.Mkdir(nextSnapshot, 0700)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			cleanuperror := os.RemoveAll(nextSnapshot)
			if cleanuperror != nil {
				log.Printf("error removing temporary snapshot creation files: %v", cleanuperror)
			}
		}
	}()

	startLogIndex := uint64(1)

	_, err = os.Stat(currentSnapshot)
	if err == nil {
		metaInfo, err := getSnapshotMetaInformation(currentSnapshot)
		if err != nil {
			return err
		}
		startLogIndex = metaInfo.LastIncludedIndex + 1

		err = startNextSnapshotWithCurrent(currentSnapshot, nextSnapshot)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("could not access current snapshot directory: %s", err)
	} else {
		err = s.startNextSnapshot(nextSnapshot)
		if err != nil {
			return err
		}
	}

	lastIncludedTerm, err := s.applyLogUpdates(nextSnapshot, startLogIndex, lastIncludedIndex)
	if err != nil {
		return err
	}

	err = tarSnapshot(nextSnapshot)
	if err != nil {
		return err
	}

	err = saveSnapshotMetaInformation(nextSnapshot, lastIncludedIndex, lastIncludedTerm, true)
	if err != nil {
		return err
	}

	err = performCleanup(nextSnapshot)
	if err != nil {
		return err
	}

	s.State.NewSnapshotCreated <- true

	return nil
}

// RevertToSnapshot revets the statemachine to the snapshot state and removes
// all log entries.
func (s *NetworkServer) RevertToSnapshot(snapshotPath string) error {
	s.State.ApplyEntryLock.Lock()
	defer s.State.ApplyEntryLock.Unlock()

	snapshotMeta, err := getSnapshotMetaInformation(snapshotPath)
	if err != nil {
		if err != nil {
			return fmt.Errorf("error reverting to snapshot: %s", err)
		}
	}

	err = libpfs.GetFileSystemLock(s.State.pfsDirectory, libpfs.ExclusiveLock)
	if err != nil {
		return fmt.Errorf("error reverting to snapshot: %s", err)
	}

	defer func() {
		err := libpfs.UnLockFileSystem(s.State.pfsDirectory)
		if err != nil {
			log.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	err = unpackTarFile(path.Join(snapshotPath, TarFileName), s.State.pfsDirectory)
	if err != nil {
		return fmt.Errorf("error reverting to snapshot: %s", err)
	}

	err = s.State.Configuration.UpdateFromConfigurationFile(
		path.Join(s.State.pfsDirectory, PersistentConfigurationFileName),
		snapshotMeta.LastIncludedIndex)
	if err != nil {
		return fmt.Errorf("error reverting to snapshot: %s", err)
	}

	err = os.Remove(path.Join(s.State.pfsDirectory, PersistentConfigurationFileName))
	if err != nil {
		log.Print("Unable to delete snapshot configuration file")
	}

	err = keyman.StateMachine.UpdateFromStateFile(path.Join(s.State.pfsDirectory, "meta", keyman.KsmFileName+"-tar"))
	if err != nil {
		return fmt.Errorf("error reverting to snapshot: %s", err)
	}

	err = os.Remove(path.Join(s.State.pfsDirectory, "meta", keyman.KsmFileName+"-tar"))
	if err != nil {
		log.Printf("Unable to delete snapshot configuration file: %v", err)
	}

	currentSnapshot := path.Join(s.raftInfoDirectory, SnapshotDirectory, CurrentSnapshotDirectory)
	if currentSnapshot != snapshotPath {
		err = os.Rename(snapshotPath, currentSnapshot)
		if err != nil {
			return fmt.Errorf("error reverting to snapshot: %s", err)
		}
	}

	s.State.Log.DiscardAllLogEntries(snapshotMeta.LastIncludedIndex, snapshotMeta.LastIncludedTerm)
	s.State.SetLastApplied(snapshotMeta.LastIncludedIndex)
	s.State.SetCommitIndex(snapshotMeta.LastIncludedIndex)

	return nil
}

// InstallSnapshot performs snapshot installation
func (s *NetworkServer) InstallSnapshot(ctx context.Context, req *pb.SnapshotRequest) (*pb.SnapshotResponse, error) {
	if req.Term < s.State.GetCurrentTerm() {
		return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, nil
	}

	snapshotPath := path.Join(s.raftInfoDirectory, SnapshotDirectory, req.LeaderId+strconv.FormatUint(req.LastIncludedIndex, 10))
	if req.Offset == 0 {
		err := os.RemoveAll(snapshotPath)
		if err != nil {
			log.Printf("Error receiving snapshot: %v", err)
			return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, err
		}

		err = os.Mkdir(snapshotPath, 0700)
		if err != nil {
			log.Printf("Error receiving snapshot: %v", err)
			return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, err
		}

		snapshotFile, err := os.Create(path.Join(snapshotPath, TarFileName))
		if err != nil {
			log.Printf("Error receiving snapshot: %v", err)
			return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, err
		}
		snapshotFile.Close()
	}

	snapshotFile, err := os.OpenFile(path.Join(snapshotPath, TarFileName), os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Error receiving snapshot: %v", err)
		return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, err
	}
	defer snapshotFile.Close()

	bytesWritten, err := snapshotFile.WriteAt(req.Data, int64(req.Offset))
	if err != nil {
		log.Printf("Error receiving snapshot: %v", err)
		return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, err
	}
	if bytesWritten != len(req.Data) {
		log.Print("Error receiving snapshot: incorrect number of bytes written to snapshot file")
		return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()},
			errors.New("incorrect number of bytes written to snapshot file")
	}

	if !req.Done {
		return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, nil
	}

	_ = saveSnapshotMetaInformation(snapshotPath, req.LastIncludedIndex, req.LastIncludedTerm, false)
	s.State.NewSnapshotCreated <- true

	return &pb.SnapshotResponse{Term: s.State.GetCurrentTerm()}, nil
}

func (s *NetworkServer) sendSnapshot(node *Node) {
	defer s.Wait.Done()
	defer s.State.DecrementSnapshotCounter()
	defer s.State.Configuration.SetSendingSnapshot(node.NodeID, false)

	conn, err := s.Dial(node, HeartbeatTimeout)
	if err != nil {
		log.Printf("error sending snapshot: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewRaftNetworkServiceClient(conn)
	currentSnapshot := path.Join(s.raftInfoDirectory, SnapshotDirectory, CurrentSnapshotDirectory)
	snapshotMeta, err := getSnapshotMetaInformation(currentSnapshot)
	if err != nil {
		log.Printf("Error sending snapshot: %v", err)
		return
	}

	snapshotFile, err := os.Open(path.Join(currentSnapshot, TarFileName))
	if err != nil {
		log.Printf("Error sending snapshot: %v", err)
		return
	}
	defer snapshotFile.Close()

	snapshotChunk := make([]byte, SnapshotChunkSize)
	snapshotFileOffset := int64(0)
	installRequestsFailed := 0

	for {
		select {
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				log.Printf("Stop sending snapshot to %s", node)
				return
			}
		default:
			if s.State.GetCurrentState() != LEADER {
				log.Print("Ceasing sending snapshot due to state change")
				return
			}

			done := false
			bytesRead, err := snapshotFile.ReadAt(snapshotChunk, snapshotFileOffset)
			if err != nil {
				if err == io.EOF {
					done = true
				} else {
					log.Printf("Error sending snapshot: %v", err)
					return
				}
			}

			response, err := client.Snapshot(context.Background(), &pb.SnapshotRequest{
				Term:              s.State.GetCurrentTerm(),
				LeaderId:          s.nodeDetails.NodeID,
				LastIncludedIndex: snapshotMeta.LastIncludedIndex,
				LastIncludedTerm:  snapshotMeta.LastIncludedTerm,
				Offset:            uint64(snapshotFileOffset),
				Data:              snapshotChunk[:bytesRead],
				Done:              done,
			})
			if err == nil {
				if response.Term > s.State.GetCurrentTerm() {
					s.State.StopLeading <- true
					return
				}
				if done {
					log.Printf("Successfully send complete snapshot to %s", node)
					s.State.Configuration.SetNextIndex(node.NodeID, snapshotMeta.LastIncludedIndex+1)
					return
				}
				snapshotFileOffset = snapshotFileOffset + int64(bytesRead)
			} else {
				if installRequestsFailed > MaxInstallSnapshotFails {
					log.Printf("InstallSnapshot request failed repeatedly: %v", err)
					return
				}
				log.Printf("InstallSnapshot request failed: %v", err)
				installRequestsFailed++
			}
		}
	}
}

//Update the current snapshot to the most recent snapshot available and remove all incomplete snapshots
func (s *NetworkServer) updateCurrentSnapshot() error {
	snapshots, err := ioutil.ReadDir(path.Join(s.raftInfoDirectory, SnapshotDirectory))
	if err != nil {
		return fmt.Errorf("unable to update current snapshot: %s", err)
	}

	currentSnapshot := path.Join(s.raftInfoDirectory, SnapshotDirectory, CurrentSnapshotDirectory)
	currentSnapshotMeta, err := getSnapshotMetaInformation(currentSnapshot)
	if err != nil {
		currentSnapshotMeta = &SnapShotInfo{
			LastIncludedIndex: 0,
			LastIncludedTerm:  0,
			SelfCreated:       false,
		}
	}

	mostRecentSnapshot := ""
	mostRecentSnapshotMeta := currentSnapshotMeta

	for i := 0; i < len(snapshots); i++ {
		snapshotPath := path.Join(s.raftInfoDirectory, SnapshotDirectory, snapshots[i].Name())
		snapshotMeta, err := getSnapshotMetaInformation(snapshotPath)
		if err != nil {
			log.Printf("error updating current snapshot: %v", err)
		} else {
			if snapshotMeta.LastIncludedIndex > mostRecentSnapshotMeta.LastIncludedIndex {
				mostRecentSnapshot = snapshotPath
				mostRecentSnapshotMeta = snapshotMeta
			}
		}
	}

	if mostRecentSnapshot != "" {
		err = os.RemoveAll(currentSnapshot)
		if err != nil {
			return fmt.Errorf("unable to update current snapshot: %s", err)
		}

		err = os.Rename(mostRecentSnapshot, currentSnapshot)
		if err != nil {
			log.Fatalf("Failed to rename new snapshot after deleteing current snapshot: %v", err)
		}

		if mostRecentSnapshotMeta.SelfCreated {
			s.State.Log.DiscardLogEntriesBefore(mostRecentSnapshotMeta.LastIncludedIndex, mostRecentSnapshotMeta.LastIncludedTerm)
		} else {
			err = s.RevertToSnapshot(currentSnapshot)
			if err != nil {
				log.Fatalf("Update current snapshot failed: %v", err)
			}
		}

		for i := 0; i < len(snapshots); i++ {
			snapshotPath := path.Join(s.raftInfoDirectory, SnapshotDirectory, snapshots[i].Name())
			if snapshotPath != currentSnapshot && snapshotPath != mostRecentSnapshot {
				err = os.RemoveAll(snapshotPath)
				if err != nil {
					log.Printf("error updating current snapshot: %v", err)
				}
			}
		}
	}

	return nil
}

func (s *NetworkServer) manageSnapshoting() {
	defer s.Wait.Done()
	snapshotTimer := time.NewTimer(SnapshortInteval)
	for {
		select {
		case _, ok := <-s.Quit:
			if !ok {
				s.QuitChannelClosed = true
				log.Print("Exiting snapshot management loop")
				return
			}
		case <-snapshotTimer.C:
			if !s.State.GetPerformingSnapshot() {
				if s.State.Log.GetLogSizeBytes() > SnapshotLogsize {
					s.Wait.Add(1)
					go func() {
						defer s.Wait.Done()
						err := s.CreateSnapshot(s.State.GetLastApplied())
						if err != nil {
							log.Printf("manage snapshotting: %v", err)
						}
					}()
				}
			}
			snapshotTimer.Reset(SnapshortInteval)
		case <-s.State.SnapshotCounterAtZero:
			err := s.updateCurrentSnapshot()
			if err != nil {
				log.Printf("manage snapshoting: %v", err)
			}
		case <-s.State.NewSnapshotCreated:
			log.Print("New Snapshot Created")
			if s.State.GetSnapshotCounterValue() == 0 {
				err := s.updateCurrentSnapshot()
				if err != nil {
					log.Printf("manage snapshoting: %v", err)
				}
			}
		case node := <-s.State.SendSnapshot:
			if !s.State.Configuration.GetSendingSnapshot(node.NodeID) {
				log.Printf("Send current snapshot to %s", node)
				s.State.Configuration.SetSendingSnapshot(node.NodeID, true)
				s.Wait.Add(1)
				s.State.IncrementSnapshotCounter()
				go s.sendSnapshot(&node)
			}
		}
	}
}
