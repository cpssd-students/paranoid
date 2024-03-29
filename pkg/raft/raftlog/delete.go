package raftlog

import (
	"errors"
	"log"
	"os"
	"path"
	"strconv"
)

func (rl *RaftLog) deleteLogEntry(index uint64) {
	entryPath := path.Join(
		rl.logDir,
		LogEntryDirectoryName,
		strconv.FormatUint(storageIndexToFileIndex(index), 10),
	)
	fi, err := os.Stat(entryPath)
	if err != nil {
		log.Fatalf("unable to delete log of index %d with error: %v", index, err)
	}
	entrySize := uint64(fi.Size())

	if err = os.Remove(entryPath); err != nil {
		log.Fatalf("unable to delete log of index %d with error: %v", index, err)
	}
	rl.setLogSizeBytes(rl.logSizeBytes - entrySize)
}

// DiscardLogEntriesAfter discards an entry in the logs at startIndex and all logs after it
func (rl *RaftLog) DiscardLogEntriesAfter(startIndex uint64) error {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()

	if startIndex <= rl.startIndex || startIndex >= rl.currentIndex {
		return errors.New("index out of bounds")
	}

	for i := rl.currentIndex - 1; i >= startIndex; i-- {
		rl.deleteLogEntry(i)
		rl.currentIndex--
	}

	if rl.currentIndex > 1 {
		logEntry, err := rl.GetLogEntryUnsafe(rl.currentIndex - 1)
		if err != nil {
			log.Fatalf("error deleting log entries: %v", err)
		}
		rl.mostRecentTerm = logEntry.Term
	} else {
		rl.mostRecentTerm = 0
	}

	return nil
}

// DiscardLogEntriesBefore discards an entry in the logs at endIndex and all logs before it
func (rl *RaftLog) DiscardLogEntriesBefore(endIndex, endTerm uint64) {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()

	for i := rl.startIndex + 1; i <= min(endIndex, rl.currentIndex-1); i++ {
		logEntry, err := rl.GetLogEntryUnsafe(i)
		if err != nil {
			log.Fatalf("error deleting log entries: %v", err)
		}

		rl.deleteLogEntry(i)
		rl.setStartIndex(i)
		rl.setStartTerm(logEntry.Term)
	}

	if rl.startIndex >= rl.currentIndex {
		rl.currentIndex = rl.startIndex + 1
		rl.mostRecentTerm = endTerm
	}
}

// DiscardAllLogEntries once a new snapshot has been reverted to
func (rl *RaftLog) DiscardAllLogEntries(snapshotIndex, snapshotTerm uint64) {
	rl.indexLock.Lock()
	defer rl.indexLock.Unlock()

	err := os.RemoveAll(path.Join(rl.logDir, LogEntryDirectoryName))
	if err != nil {
		log.Fatalf("error deleting all log entries: %v", err)
	}

	err = os.Mkdir(path.Join(rl.logDir, LogEntryDirectoryName), 0700)
	if err != nil {
		log.Fatalf("error deleting all log entries: %v", err)
	}

	rl.setLogSizeBytes(0)
	rl.setStartIndex(snapshotIndex)
	rl.setStartTerm(snapshotTerm)
	rl.currentIndex = rl.startIndex + 1
	rl.mostRecentTerm = snapshotTerm
}
