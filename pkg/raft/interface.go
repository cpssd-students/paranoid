package raft

// This file contains the list of interfaces used by raft

import (
	"errors"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
	"github.com/cpssd-students/paranoid/pkg/libpfs"
	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
	pb "github.com/cpssd-students/paranoid/proto/raft"
)

// ActionType is the base type of the different action
type ActionType uint32

// List of the ActionTypes
const (
	TypeWrite ActionType = iota
	TypeCreat
	TypeChmod
	TypeTruncate
	TypeUtimes
	TypeRename
	TypeLink
	TypeSymlink
	TypeUnlink
	TypeMkdir
	TypeRmdir
)

// StateMachineResult containing the data of the state machine
type StateMachineResult struct {
	Code         returncodes.Code
	Err          error
	BytesWritten int
	KSMResult    *ksmResult
}

type ksmResult struct {
	GenerationNumber int
	Peers            []string
}

// EntryAppliedInfo stores the information with the state machine results
type EntryAppliedInfo struct {
	Index  uint64
	Result *StateMachineResult
}

//StartRaft server given a listener, node information a directory to store information
//Only used for testing purposes
func StartRaft(lis *net.Listener, nodeDetails Node, pfsDirectory, raftInfoDirectory string,
	startConfiguration *StartConfiguration) (*NetworkServer, *grpc.Server) {

	var opts []grpc.ServerOption
	srv := grpc.NewServer(opts...)
	raftServer := NewNetworkServer(nodeDetails, pfsDirectory, raftInfoDirectory, startConfiguration, false, false, false)
	pb.RegisterRaftNetworkServer(srv, raftServer)
	raftServer.Wait.Add(1)
	go func() {
		Log.Info("RaftNetworkServer started")
		err := srv.Serve(*lis)
		if err != nil {
			Log.Error("Error running RaftNetworkServer", err)
		}
	}()
	return raftServer, srv
}

// RequestAddLogEntry from a client. If the mode is not the leader, it must
// follow the request to the leader. Only returns once the request has been
// committed to the State machine
func (s *NetworkServer) RequestAddLogEntry(entry *pb.Entry) (*StateMachineResult, error) {
	s.addEntryLock.Lock()
	defer s.addEntryLock.Unlock()
	currentState := s.State.GetCurrentState()

	s.State.SetWaitingForApplied(true)
	defer s.State.SetWaitingForApplied(false)

	//Add entry to leaders Log
	if NodeType(currentState) == LEADER {
		err := s.addLogEntryLeader(entry)
		if err != nil {
			return nil, err
		}
	} else if NodeType(currentState) == FOLLOWER {
		if s.State.GetLeaderID() != "" {
			err := s.sendLeaderLogEntry(entry)
			if err != nil {
				return nil, err
			}
		} else {
			select {
			case <-time.After(20 * time.Second):
				return nil, errors.New("could not find a leader")
			case <-s.State.LeaderElected:
				if s.State.GetCurrentState() == LEADER {
					err := s.addLogEntryLeader(entry)
					if err != nil {
						return nil, err
					}
				} else {
					err := s.sendLeaderLogEntry(entry)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	} else {
		count := 0
		for {
			count++
			if count > 40 {
				return nil, errors.New("could not find a leader")
			}
			time.Sleep(500 * time.Millisecond)
			currentState = s.State.GetCurrentState()
			if currentState != CANDIDATE {
				break
			}
		}
		if currentState == LEADER {
			err := s.addLogEntryLeader(entry)
			if err != nil {
				return nil, err
			}
		} else {
			err := s.sendLeaderLogEntry(entry)
			if err != nil {
				return nil, err
			}
		}
	}

	//Wait for the Log entry to be applied
	timer := time.NewTimer(EntryAppliedTimeout)
	for {
		select {
		case <-timer.C:
			return nil, errors.New("waited too long to commit Log entry")
		case appliedEntry := <-s.State.EntryApplied:
			LogEntry, err := s.State.Log.GetLogEntry(appliedEntry.Index)
			if err != nil {
				Log.Fatal("unable to get log entry:", err)
			}
			if LogEntry.Entry.Uuid == entry.Uuid {
				return appliedEntry.Result, nil
			}
		}
	}
	// return nil, errors.New("waited too long to commit Log entry")
}

// RequestKeyStateUpdate requests and update
func (s *NetworkServer) RequestKeyStateUpdate(owner, holder *pb.Node, generation int64) error {
	entry := &pb.Entry{
		Type: pb.Entry_KeyStateCommand,
		Uuid: generateNewUUID(),
		KeyCommand: &pb.KeyStateCommand{
			Type:       pb.KeyStateCommand_UpdateKeyPiece,
			KeyOwner:   owner,
			KeyHolder:  holder,
			Generation: generation,
		},
	}
	result, err := s.RequestAddLogEntry(entry)
	if err != nil {
		Log.Error("failed to add log entry for key state update:", err)
		return err
	}
	return result.Err
}

// RequestNewGeneration retrns a number, a list of peer nodes, and an error.
func (s *NetworkServer) RequestNewGeneration(newNode string) (int, []string, error) {
	entry := &pb.Entry{
		Type: pb.Entry_KeyStateCommand,
		Uuid: generateNewUUID(),
		KeyCommand: &pb.KeyStateCommand{
			Type:    pb.KeyStateCommand_NewGeneration,
			NewNode: newNode,
		},
	}
	result, err := s.RequestAddLogEntry(entry)
	if err != nil {
		Log.Error("failed to add log entry for new generation:", err)
		return -1, nil, err
	}
	return result.KSMResult.GenerationNumber, result.KSMResult.Peers, result.Err
}

// RequestOwnerComplete of the node
func (s *NetworkServer) RequestOwnerComplete(nodeID string, generation int64) error {
	entry := &pb.Entry{
		Type: pb.Entry_KeyStateCommand,
		Uuid: generateNewUUID(),
		KeyCommand: &pb.KeyStateCommand{
			Type:          pb.KeyStateCommand_OwnerComplete,
			OwnerComplete: nodeID,
			Generation:    generation,
		},
	}
	result, err := s.RequestAddLogEntry(entry)
	if err != nil {
		Log.Error("failed to add log entry for owner complete:", err)
		return err
	}
	return result.Err
}

// RequestWriteCommand performs a write
func (s *NetworkServer) RequestWriteCommand(filePath string, offset, length int64, data []byte) (returnCode returncodes.Code, bytesWrote int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:   uint32(TypeWrite),
			Path:   filePath,
			Data:   data,
			Offset: offset,
			Length: length,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, 0, err
	}
	return stateMachineResult.Code, stateMachineResult.BytesWritten, stateMachineResult.Err
}

// RequestCreatCommand performs the Creat command
func (s *NetworkServer) RequestCreatCommand(filePath string, mode uint32) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: uint32(TypeCreat),
			Path: filePath,
			Mode: mode,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestChmodCommand performs the Chmod command
func (s *NetworkServer) RequestChmodCommand(filePath string, mode uint32) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: uint32(TypeChmod),
			Path: filePath,
			Mode: mode,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestTruncateCommand performs the Truncate command
func (s *NetworkServer) RequestTruncateCommand(filePath string, length int64) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:   uint32(TypeTruncate),
			Path:   filePath,
			Length: length,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func splitTime(t *time.Time) (int64, int64) {
	if t != nil {
		return int64(t.Second()), int64(t.Nanosecond())
	}
	return 0, 0
}

// RequestUtimesCommand performs the Utimes command
func (s *NetworkServer) RequestUtimesCommand(filePath string, atime, mtime *time.Time) (returnCode returncodes.Code, returnError error) {
	accessSeconds, accessNanoSeconds := splitTime(atime)
	modifySeconds, modifyNanoSeconds := splitTime(mtime)

	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:              uint32(TypeUtimes),
			Path:              filePath,
			AccessSeconds:     accessSeconds,
			AccessNanoseconds: accessNanoSeconds,
			ModifySeconds:     modifySeconds,
			ModifyNanoseconds: modifyNanoSeconds,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestRenameCommand performs the rename command
func (s *NetworkServer) RequestRenameCommand(oldPath, newPath string) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:    uint32(TypeRename),
			OldPath: oldPath,
			NewPath: newPath,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestLinkCommand performs the link command
func (s *NetworkServer) RequestLinkCommand(oldPath, newPath string) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:    uint32(TypeLink),
			OldPath: oldPath,
			NewPath: newPath,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestSymlinkCommand performs the symlink command
func (s *NetworkServer) RequestSymlinkCommand(oldPath, newPath string) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:    uint32(TypeSymlink),
			OldPath: oldPath,
			NewPath: newPath,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestUnlinkCommand performs the unlink command
func (s *NetworkServer) RequestUnlinkCommand(filePath string) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: uint32(TypeUnlink),
			Path: filePath,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestMkdirCommand performs the mkdir command
func (s *NetworkServer) RequestMkdirCommand(filePath string, mode uint32) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: uint32(TypeMkdir),
			Path: filePath,
			Mode: mode,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestRmdirCommand performs the rmdir command
func (s *NetworkServer) RequestRmdirCommand(filePath string) (returnCode returncodes.Code, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: uint32(TypeRmdir),
			Path: filePath,
		},
	}
	stateMachineResult, err := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

// RequestChangeConfiguration performs a change in Configuration
func (s *NetworkServer) RequestChangeConfiguration(nodes []Node) error {
	Log.Info("Configuration change requested:", nodes)
	entry := &pb.Entry{
		Type: pb.Entry_ConfigurationChange,
		Uuid: generateNewUUID(),
		Config: &pb.Configuration{
			Type:  pb.Configuration_FutureConfiguration,
			Nodes: convertNodesToProto(nodes),
		},
	}
	_, err := s.RequestAddLogEntry(entry)
	return err
}

// RequestAddNodeToConfiguration adds a node to configuration
func (s *NetworkServer) RequestAddNodeToConfiguration(node Node) error {
	if s.State.Configuration.InConfiguration(node.NodeID) {
		return nil
	}
	nodes := append(s.State.Configuration.GetNodesList(), node)
	return s.RequestChangeConfiguration(nodes)
}

//ChangeNodeLocation changes the IP and Port of a given node
func (s *NetworkServer) ChangeNodeLocation(UUID, IP, Port string) {
	s.State.Configuration.ChangeNodeLocation(UUID, IP, Port)
}

// PerformLibPfsCommand performs a libpfs command
func PerformLibPfsCommand(directory string, command *pb.StateMachineCommand) *StateMachineResult {
	switch ActionType(command.Type) {
	case TypeWrite:
		code, bytesWritten, err := libpfs.WriteCommand(directory, command.Path, int64(command.Offset), int64(command.Length), command.Data)
		return &StateMachineResult{Code: code, Err: err, BytesWritten: bytesWritten}
	case TypeCreat:
		code, err := libpfs.CreatCommand(directory, command.Path, os.FileMode(command.Mode))
		return &StateMachineResult{Code: code, Err: err}
	case TypeChmod:
		code, err := libpfs.ChmodCommand(directory, command.Path, os.FileMode(command.Mode))
		return &StateMachineResult{Code: code, Err: err}
	case TypeTruncate:
		code, err := libpfs.TruncateCommand(directory, command.Path, int64(command.Length))
		return &StateMachineResult{Code: code, Err: err}
	case TypeUtimes:
		var atime *time.Time
		var mtime *time.Time
		if command.AccessNanoseconds != 0 || command.AccessSeconds != 0 {
			time := time.Unix(command.AccessSeconds, command.AccessNanoseconds)
			atime = &time
		}
		if command.ModifyNanoseconds != 0 || command.ModifySeconds != 0 {
			time := time.Unix(command.ModifySeconds, command.ModifyNanoseconds)
			mtime = &time
		}
		code, err := libpfs.UtimesCommand(directory, command.Path, atime, mtime)
		return &StateMachineResult{Code: code, Err: err}
	case TypeRename:
		code, err := libpfs.RenameCommand(directory, command.OldPath, command.NewPath)
		return &StateMachineResult{Code: code, Err: err}
	case TypeLink:
		code, err := libpfs.LinkCommand(directory, command.OldPath, command.NewPath)
		return &StateMachineResult{Code: code, Err: err}
	case TypeSymlink:
		code, err := libpfs.SymlinkCommand(directory, command.OldPath, command.NewPath)
		return &StateMachineResult{Code: code, Err: err}
	case TypeUnlink:
		code, err := libpfs.UnlinkCommand(directory, command.Path)
		return &StateMachineResult{Code: code, Err: err}
	case TypeMkdir:
		code, err := libpfs.MkdirCommand(directory, command.Path, os.FileMode(command.Mode))
		return &StateMachineResult{Code: code, Err: err}
	case TypeRmdir:
		code, err := libpfs.RmdirCommand(directory, command.Path)
		return &StateMachineResult{Code: code, Err: err}
	}
	Log.Fatal("Unrecognised command type")
	return nil
}

// PerformKSMCommand updates the keys
func PerformKSMCommand(sateMachine *keyman.KeyStateMachine, keyCommand *pb.KeyStateCommand) *StateMachineResult {
	switch keyCommand.Type {
	case pb.KeyStateCommand_UpdateKeyPiece:
		err := keyman.StateMachine.Update(keyCommand)
		return &StateMachineResult{
			Err: err,
		}
	case pb.KeyStateCommand_NewGeneration:
		generationNumber, peers, err := keyman.StateMachine.NewGeneration(keyCommand.NewNode)
		return &StateMachineResult{
			Err: err,
			KSMResult: &ksmResult{
				GenerationNumber: int(generationNumber),
				Peers:            peers,
			},
		}
	case pb.KeyStateCommand_OwnerComplete:
		err := keyman.StateMachine.OwnerComplete(keyCommand.OwnerComplete, keyCommand.Generation)
		return &StateMachineResult{
			Err: err,
		}
	}
	Log.Fatalf("Unrecognised command type: %s", keyCommand.Type)
	return nil
}
