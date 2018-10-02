package libpfs

import (
	"errors"
	"fmt"

	"paranoid/libpfs/returncodes"
	log "paranoid/logger"
)

//AccessCommand is used by fuse to check if it has access to a given file.
func AccessCommand(paranoidDirectory, filePath string, mode uint32) (returnCode returncodes.Code, returnError error) {
	log.V(1).Info("access called on %s in %s", filePath, paranoidDirectory)

	err := GetFileSystemLock(paranoidDirectory, SharedLock)
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

	namePath := getParanoidPath(paranoidDirectory, filePath)

	fileType, err := getFileType(paranoidDirectory, namePath)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	if fileType == typeENOENT {
		return returncodes.ENOENT, errors.New(filePath + " does not exist")
	}

	inodeNameBytes, code, err := getFileInode(namePath)
	if code != returncodes.OK || err != nil {
		return code, err
	}

	inodeName := string(inodeNameBytes)

	code, err = canAccessFile(paranoidDirectory, inodeName, mode)
	if err != nil {
		return code, fmt.Errorf("unable to access %s: %s", filePath, err)
	}
	return returncodes.OK, nil
}
