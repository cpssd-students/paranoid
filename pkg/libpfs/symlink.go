package libpfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"syscall"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

// SymlinkCommand creates a symbolic link
func SymlinkCommand(
	paranoidDirectory, existingFilePath, targetFilePath string,
) (returnCode returncodes.Code, returnError error) {
	log.Printf("symlinking %s with %s in %s",
		existingFilePath, targetFilePath, paranoidDirectory)

	targetParanoidPath := getParanoidPath(paranoidDirectory, targetFilePath)

	err := GetFileSystemLock(paranoidDirectory, ExclusiveLock)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	defer func() {
		err := UnLockFileSystem(paranoidDirectory)
		if err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
		}
	}()

	targetFilePathType, err := getFileType(paranoidDirectory, targetParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	if targetFilePathType != typeENOENT {
		return returncodes.EEXIST, errors.New(targetFilePath + " already exists")
	}

	uuidBytes, err := generateNewInode()
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	uuidString := string(uuidBytes)

	err = ioutil.WriteFile(targetParanoidPath, uuidBytes, 0600)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error writing file: %s", err)
	}

	contentsFile := path.Join(paranoidDirectory, "contents", uuidString)
	err = os.Symlink(os.DevNull, contentsFile)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error creating symlinks: %s", err)
	}

	fi, err := os.Lstat(contentsFile)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error Lstating file: %s", err)
	}

	stat := fi.Sys().(*syscall.Stat_t)

	nodeData := &inode{
		Mode:  os.FileMode(stat.Mode),
		Inode: uuidString,
		Count: 1,
		Link:  existingFilePath,
	}

	jsonData, err := json.Marshal(nodeData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error marshalling json: %s", err)
	}

	err = ioutil.WriteFile(path.Join(paranoidDirectory, "inodes", uuidString), jsonData, 0600)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error writing inodes file: %s", err)
	}

	return returncodes.OK, nil
}
