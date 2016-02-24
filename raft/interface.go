//This file stores functions used to interface in or out of raft
package raft

import (
	"errors"
	"github.com/cpssd/paranoid/libpfs/commands"
	"github.com/cpssd/paranoid/libpfs/returncodes"
	pb "github.com/cpssd/paranoid/proto/raft"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

const (
	TYPE_WRITE uint32 = iota
	TYPE_CREAT
	TYPE_CHMOD
	TYPE_TRUNCATE
	TYPE_UTIMES
	TYPE_RENAME
	TYPE_LINK
	TYPE_SYMLINK
	TYPE_UNLINK
	TYPE_MKDIR
	TYPE_RMDIR
)

type StateMachineResult struct {
	Code         int
	Err          error
	BytesWritten int
}

type EntryAppliedInfo struct {
	Index  uint64
	Result *StateMachineResult
}

//Starts a raft server given a listener, node information a directory to store information
//A start configuration can be given for testing or for the first node in a cluster
func StartRaft(lis *net.Listener, nodeDetails Node, pfsDirectory, raftInfoDirectory string,
	startConfiguration *StartConfiguration) (*RaftNetworkServer, *grpc.Server) {

	var opts []grpc.ServerOption
	srv := grpc.NewServer(opts...)
	raftServer := NewRaftNetworkServer(nodeDetails, pfsDirectory, raftInfoDirectory, startConfiguration)
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

//A request to add a Log entry from a client. If the node is not the leader, it must forward the request to the leader.
//Only return once the request has been commited to the State machine
func (s *RaftNetworkServer) RequestAddLogEntry(entry *pb.Entry) (error, *StateMachineResult) {
	s.addEntryLock.Lock()
	defer s.addEntryLock.Unlock()
	currentState := s.State.GetCurrentState()

	s.State.SetWaitingForApplied(true)
	defer s.State.SetWaitingForApplied(false)

	//Add entry to leaders Log
	if currentState == LEADER {
		err := s.addLogEntryLeader(entry)
		if err != nil {
			return err, nil
		}
	} else if currentState == FOLLOWER {
		if s.State.GetLeaderId() != "" {
			err := s.sendLeaderLogEntry(entry)
			if err != nil {
				return err, nil
			}
		} else {
			select {
			case <-time.After(20 * time.Second):
				return errors.New("Could not find a leader"), nil
			case <-s.State.LeaderElected:
				if s.State.GetCurrentState() == LEADER {
					err := s.addLogEntryLeader(entry)
					if err != nil {
						return err, nil
					}
				} else {
					err := s.sendLeaderLogEntry(entry)
					if err != nil {
						return err, nil
					}
				}
			}
		}
	} else {
		count := 0
		for {
			count++
			if count > 40 {
				return errors.New("Could not find a leader"), nil
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
				return err, nil
			}
		} else {
			err := s.sendLeaderLogEntry(entry)
			if err != nil {
				return err, nil
			}
		}
	}

	//Wait for the Log entry to be applied
	timer := time.NewTimer(ENTRY_APPLIED_TIMEOUT)
	for {
		select {
		case <-timer.C:
			return errors.New("Waited too long to commit Log entry"), nil
		case appliedEntry := <-s.State.EntryApplied:
			LogEntry, err := s.State.Log.GetLogEntry(appliedEntry.Index)
			if err != nil {
				Log.Fatal("Unable to get log entry:", err)
			}
			if LogEntry.Entry.Uuid == entry.Uuid {
				return nil, appliedEntry.Result
			}
		}
	}
	return errors.New("Waited too long to commit Log entry"), nil
}

func (s *RaftNetworkServer) RequestWriteCommand(filePath string, offset, length uint64,
	data []byte) (returnCode int, returnError error, bytesWrote int) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:   TYPE_WRITE,
			Path:   filePath,
			Data:   data,
			Offset: offset,
			Length: length,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err, 0
	}
	return stateMachineResult.Code, stateMachineResult.Err, stateMachineResult.BytesWritten
}

func (s *RaftNetworkServer) RequestCreatCommand(filePath string, mode uint32) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: TYPE_CREAT,
			Path: filePath,
			Mode: mode,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestChmodCommand(filePath string, mode uint32) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: TYPE_CHMOD,
			Path: filePath,
			Mode: mode,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestTruncateCommand(filePath string, length uint64) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:   TYPE_TRUNCATE,
			Path:   filePath,
			Length: length,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestUtimesCommand(filePath string, accessSeconds, accessNanoSeconds, modifySeconds,
	modifyNanoSeconds int64) (returnCode int, returnError error) {

	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:              TYPE_UTIMES,
			Path:              filePath,
			AccessSeconds:     accessSeconds,
			AccessNanoseconds: accessNanoSeconds,
			ModifySeconds:     modifySeconds,
			ModifyNanoseconds: modifyNanoSeconds,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestRenameCommand(oldPath, newPath string) (returnCode int, returnError error) {

	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:    TYPE_RENAME,
			OldPath: oldPath,
			NewPath: newPath,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestLinkCommand(oldPath, newPath string) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:    TYPE_LINK,
			OldPath: oldPath,
			NewPath: newPath,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestSymlinkCommand(oldPath, newPath string) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type:    TYPE_SYMLINK,
			OldPath: oldPath,
			NewPath: newPath,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestUnlinkCommand(filePath string) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: TYPE_UNLINK,
			Path: filePath,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestMkdirCommand(filePath string, mode uint32) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: TYPE_MKDIR,
			Path: filePath,
			Mode: mode,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestRmdirCommand(filePath string) (returnCode int, returnError error) {
	entry := &pb.Entry{
		Type: pb.Entry_StateMachineCommand,
		Uuid: generateNewUUID(),
		Command: &pb.StateMachineCommand{
			Type: TYPE_RMDIR,
			Path: filePath,
		},
	}
	err, stateMachineResult := s.RequestAddLogEntry(entry)
	if err != nil {
		return returncodes.EBUSY, err
	}
	return stateMachineResult.Code, stateMachineResult.Err
}

func (s *RaftNetworkServer) RequestChangeConfiguration(nodes []Node) error {
	Log.Info("Configuration change requested:", nodes)
	entry := &pb.Entry{
		Type: pb.Entry_ConfigurationChange,
		Uuid: generateNewUUID(),
		Config: &pb.Configuration{
			Type:  pb.Configuration_FutureConfiguration,
			Nodes: convertNodesToProto(nodes),
		},
	}
	Log.Info("Nodes:", entry.Config.Nodes)
	err, _ := s.RequestAddLogEntry(entry)
	return err
}

func (s *RaftNetworkServer) RequestAddNodeToConfiguration(node Node) error {
	nodes := append(s.State.Configuration.GetNodesList(), node)
	return s.RequestChangeConfiguration(nodes)
}

func performLibPfsCommand(directory string, command *pb.StateMachineCommand) *StateMachineResult {
	switch command.Type {
	case TYPE_WRITE:
		code, err, bytesWritten := commands.WriteCommand(directory, command.Path, int64(command.Offset), int64(command.Length), command.Data)
		return &StateMachineResult{code, err, bytesWritten}
	case TYPE_CREAT:
		code, err := commands.CreatCommand(directory, command.Path, os.FileMode(command.Mode))
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_CHMOD:
		code, err := commands.ChmodCommand(directory, command.Path, os.FileMode(command.Mode))
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_TRUNCATE:
		code, err := commands.TruncateCommand(directory, command.Path, int64(command.Length))
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_UTIMES:
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
		code, err := commands.UtimesCommand(directory, command.Path, atime, mtime)
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_RENAME:
		code, err := commands.RenameCommand(directory, command.OldPath, command.NewPath)
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_LINK:
		code, err := commands.LinkCommand(directory, command.OldPath, command.NewPath)
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_SYMLINK:
		code, err := commands.SymlinkCommand(directory, command.OldPath, command.NewPath)
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_UNLINK:
		code, err := commands.UnlinkCommand(directory, command.Path)
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_MKDIR:
		code, err := commands.MkdirCommand(directory, command.Path, os.FileMode(command.Mode))
		return &StateMachineResult{Code: code, Err: err}
	case TYPE_RMDIR:
		code, err := commands.RmdirCommand(directory, command.Path)
		return &StateMachineResult{Code: code, Err: err}
	}
	Log.Fatal("Unrecognised command type")
	return nil
}
