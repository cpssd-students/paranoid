package libpfs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pp2p/paranoid/libpfs/returncodes"
	log "github.com/pp2p/paranoid/logger"
)

// RenameCommand is called when renaming a file
func RenameCommand(paranoidDirectory, oldFilePath, newFilePath string) (returnCode returncodes.Code, returnError error) {
	log.V(1).Info("rename %s to %s in %s",
		oldFilePath, newFilePath, paranoidDirectory)

	oldFileParanoidPath := getParanoidPath(paranoidDirectory, oldFilePath)
	newFileParanoidPath := getParanoidPath(paranoidDirectory, newFilePath)

	oldFileType, err := getFileType(paranoidDirectory, oldFileParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	newFileType, err := getFileType(paranoidDirectory, newFileParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	err = GetFileSystemLock(paranoidDirectory, ExclusiveLock)
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

	if oldFileType == typeENOENT {
		return returncodes.ENOENT, errors.New(oldFilePath + " does not exist")
	}

	if newFileType != typeENOENT {
		//Renaming is allowed when a file already exists, unless the existing file is a non empty paranoidDirectory
		if newFileType == typeFile {
			_, err := UnlinkCommand(paranoidDirectory, newFilePath)
			if err != nil {
				return returncodes.EEXIST, errors.New(newFilePath + " already exists")
			}
		} else if newFileType == typeDir {
			dirpath := getParanoidPath(paranoidDirectory, newFilePath)
			files, err := ioutil.ReadDir(dirpath)
			if err != nil || len(files) > 0 {
				return returncodes.ENOTEMPTY, errors.New(newFilePath + " is not empty")
			}
			_, err = RmdirCommand(paranoidDirectory, newFilePath)
			if err != nil {
				return returncodes.EEXIST, errors.New(newFilePath + " already exists")
			}
		}
	}

	err = os.Rename(oldFileParanoidPath, newFileParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error renaming file: %s", err)
	}

	return returncodes.OK, nil
}
