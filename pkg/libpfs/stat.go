package libpfs

import (
	"fmt"
	"log"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

// StatInfo contains the file metadata
type StatInfo struct {
	Length int64
	Ctime  time.Time
	Mtime  time.Time
	Atime  time.Time
	Mode   os.FileMode
}

// StatCommand returns information about a file as StatInfo object
func StatCommand(
	paranoidDirectory, filePath string,
) (returnCode returncodes.Code, info StatInfo, returnError error) {
	log.Printf("stat called on %s in %s", filePath, paranoidDirectory)

	if err := GetFileSystemLock(paranoidDirectory, SharedLock); err != nil {
		return returncodes.EUNEXPECTED, StatInfo{}, err
	}

	defer func() {
		if err := UnLockFileSystem(paranoidDirectory); err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
			info = StatInfo{}
		}
	}()

	namepath := getParanoidPath(paranoidDirectory, filePath)
	namePathType, err := getFileType(paranoidDirectory, namepath)
	if err != nil {
		return returncodes.EUNEXPECTED, StatInfo{}, err
	}

	if namePathType == typeENOENT {
		return returncodes.ENOENT, StatInfo{}, fmt.Errorf("%s does not exist", filePath)
	}

	inodeBytes, code, err := getFileInode(namepath)
	if code != returncodes.OK {
		return code, StatInfo{}, err
	}

	inodeName := string(inodeBytes)
	contentsFilePath := path.Join(paranoidDirectory, "contents", inodeName)

	contentsFile, err := os.Open(contentsFilePath)
	if err != nil {
		return returncodes.EUNEXPECTED, StatInfo{},
			fmt.Errorf("error opening contents file: %w", err)
	}

	fi, err := os.Lstat(contentsFilePath)
	if err != nil {
		return returncodes.EUNEXPECTED, StatInfo{},
			fmt.Errorf("error Lstating file: %w", err)
	}

	stat := fi.Sys().(*syscall.Stat_t)
	atime := lastAccess(stat)
	ctime := createTime(stat)
	mode, err := getFileMode(paranoidDirectory, inodeName)
	if err != nil {
		return returncodes.EUNEXPECTED, StatInfo{},
			fmt.Errorf("error getting filemode: %w", err)
	}

	fileLength, err := getFileLength(contentsFile)
	if err != nil {
		return returncodes.EUNEXPECTED, StatInfo{},
			fmt.Errorf("error getting file length: %w", err)
	}

	statData := &StatInfo{
		Length: fileLength,
		Mtime:  fi.ModTime(),
		Ctime:  ctime,
		Atime:  atime,
		Mode:   mode}

	return returncodes.OK, *statData, nil
}
