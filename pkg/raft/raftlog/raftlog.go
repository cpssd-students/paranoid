package raftlog

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
)

// Constants used by the raftlog
const (
	LogEntryDirectoryName = "log_entries"
	LogMetaFileName       = "logmetainfo"
)

// ErrIndexBelowStartIndex is returned if the given index is below start index.
var ErrIndexBelowStartIndex = errors.New("given index is below start index")

// PersistentLogState stores the state information
type PersistentLogState struct {
	LogSizeBytes uint64 `json:"logsizebytes"`
	StartIndex   uint64 `json:"startindex"`
	StartTerm    uint64 `json:"startterm"`
}

// RaftLog is the structure through which logging functinality can be accessed
type RaftLog struct {
	logDir         string
	startIndex     uint64
	startTerm      uint64
	logSizeBytes   uint64
	currentIndex   uint64
	mostRecentTerm uint64
	indexLock      sync.Mutex
}

// New returns an initiated instance of RaftLog
func New(logDirectory string) *RaftLog {
	rl := &RaftLog{
		logDir: logDirectory,
	}

	logEntryDirectory := path.Join(rl.logDir, LogEntryDirectoryName)
	logMetaFile := path.Join(rl.logDir, LogMetaFileName)

	fileDescriptors, err := ioutil.ReadDir(logEntryDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(logEntryDirectory, 0700); err != nil {
				log.Fatalf("failed to create log directory: %v", err)
			}
		} else if os.IsPermission(err) {
			log.Fatalf("raft logger does not have permissions for: %v", logEntryDirectory)
		} else {
			log.Fatalf("unable to read log directory: %v", err)
		}
	}

	rl.startIndex = 0
	rl.startTerm = 0
	rl.logSizeBytes = 0
	rl.mostRecentTerm = 0
	metaFileContents, err := ioutil.ReadFile(logMetaFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("unable to read raft log meta information: %v", err)
		}
	} else {
		metaInfo := &PersistentLogState{}
		err = json.Unmarshal(metaFileContents, metaInfo)
		if err != nil {
			log.Fatalf("unable to read raft log meta information: %v", err)
		}
		rl.startIndex = metaInfo.StartIndex
		rl.logSizeBytes = metaInfo.LogSizeBytes
		rl.startTerm = metaInfo.StartTerm
		rl.mostRecentTerm = metaInfo.StartTerm
	}

	rl.currentIndex = uint64(len(fileDescriptors)) + rl.startIndex + 1
	if rl.currentIndex > rl.startIndex+1 {
		logEntry, err := rl.GetLogEntry(rl.currentIndex - 1)
		if err != nil {
			log.Fatalf("failed setting up raft logger, could not get most recent term: %v", err)
		}
		rl.mostRecentTerm = logEntry.Term
	}
	return rl
}

func (rl *RaftLog) saveMetaInfo() {
	metaInfo := &PersistentLogState{
		LogSizeBytes: rl.logSizeBytes,
		StartIndex:   rl.startIndex,
		StartTerm:    rl.startTerm,
	}

	metaInfoJSON, err := json.Marshal(metaInfo)
	if err != nil {
		log.Fatalf("unable to save raft log meta information: %v", err)
	}

	newMetaFile := path.Join(rl.logDir, LogMetaFileName+"-new")
	err = ioutil.WriteFile(newMetaFile, metaInfoJSON, 0600)
	if err != nil {
		log.Fatalf("unable to save raft log meta information: %v", err)
	}

	err = os.Rename(newMetaFile, path.Join(rl.logDir, LogMetaFileName))
	if err != nil {
		log.Fatalf("unable to save raft log meta information: %v", err)
	}
}

// GetMostRecentIndex returns the index of the last log entry
func (rl *RaftLog) GetMostRecentIndex() uint64 {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()
	return rl.currentIndex - 1
}

// GetMostRecentTerm returns the term of the last log entry
func (rl *RaftLog) GetMostRecentTerm() uint64 {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()
	return rl.mostRecentTerm
}

// fileIndexToStorageIndex converts a fileIndex to a the index it is stored at
func fileIndexToStorageIndex(i uint64) uint64 {
	return i - 1000000
}

// storageIndexToFileIndex converts a storage index to a fileIndex
func storageIndexToFileIndex(i uint64) uint64 {
	return i + 1000000
}

func (rl *RaftLog) setLogSizeBytes(x uint64) {
	rl.logSizeBytes = x
	rl.saveMetaInfo()
}

// GetLogSizeBytes returns the side of the log in bytes
func (rl *RaftLog) GetLogSizeBytes() uint64 {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()
	return rl.logSizeBytes
}

func (rl *RaftLog) setStartIndex(x uint64) {
	rl.startIndex = x
	rl.saveMetaInfo()
}

func (rl *RaftLog) setStartTerm(x uint64) {
	rl.startTerm = x
	rl.saveMetaInfo()
}

// GetStartTerm returns start term
func (rl *RaftLog) GetStartTerm() uint64 {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()
	return rl.startTerm
}

// GetStartIndex returns the starting index
func (rl *RaftLog) GetStartIndex() uint64 {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()
	return rl.startIndex
}
