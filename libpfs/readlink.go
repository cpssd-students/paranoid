package libpfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"paranoid/libpfs/returncodes"
	log "paranoid/logger"
)

// ReadlinkCommand reads the value of the symbolic link
func ReadlinkCommand(paranoidDirectory, filePath string) (returnCode returncodes.Code, linkContents string, returnError error) {
	log.V(1).Info("readlink called on %s in %s", filePath, paranoidDirectory)

	err := GetFileSystemLock(paranoidDirectory, SharedLock)
	if err != nil {
		return returncodes.EUNEXPECTED, "", err
	}

	defer func() {
		err := UnLockFileSystem(paranoidDirectory)
		if err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
			linkContents = ""
		}
	}()

	link := getParanoidPath(paranoidDirectory, filePath)
	fileType, err := getFileType(paranoidDirectory, link)
	if err != nil {
		return returncodes.EUNEXPECTED, "", err
	}

	if fileType == typeENOENT {
		return returncodes.ENOENT, "", fmt.Errorf("%s does not exist", filePath)
	}

	if fileType == typeDir {
		return returncodes.EISDIR, "", fmt.Errorf("%s is a paranoidDirectory", filePath)
	}

	if fileType == typeFile {
		return returncodes.EIO, "", fmt.Errorf("%s is a file", err)
	}

	linkInode, code, err := getFileInode(link)
	if code != returncodes.OK || err != nil {
		return code, "", err
	}

	err = getFileLock(paranoidDirectory, string(linkInode), SharedLock)
	if err != nil {
		return returncodes.EUNEXPECTED, "", err
	}

	defer func() {
		err := unLockFile(paranoidDirectory, string(linkInode))
		if err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
			linkContents = ""
		}
	}()

	inodePath := path.Join(paranoidDirectory, "inodes", string(linkInode))

	inodeContents, err := ioutil.ReadFile(inodePath)
	if err != nil {
		return returncodes.EUNEXPECTED, "", fmt.Errorf("error reading link: %s", err)
	}

	inodeData := &inode{}
	err = json.Unmarshal(inodeContents, &inodeData)
	if err != nil {
		return returncodes.EUNEXPECTED, "", fmt.Errorf("error unmarshalling json: %s", err)
	}

	return returncodes.OK, inodeData.Link, nil
}
