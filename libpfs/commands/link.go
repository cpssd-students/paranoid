package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pp2p/paranoid/libpfs/returncodes"
	log "github.com/pp2p/paranoid/logger"
)

// LinkCommand creates a link of a file.
func LinkCommand(paranoidDirectory, existingFilePath, targetFilePath string) (returnCode returncodes.Code, returnError error) {
	log.V(1).Info("linking %s with %s in %s",
		existingFilePath, targetFilePath, paranoidDirectory)

	existingParanoidPath := getParanoidPath(paranoidDirectory, existingFilePath)
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

	existingFileType, err := getFileType(paranoidDirectory, existingParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	if existingFileType == typeENOENT {
		return returncodes.ENOENT, errors.New("existing file " + existingFilePath + " does not exist")
	}

	if existingFileType == typeDir {
		return returncodes.EISDIR, errors.New("existing file " + existingFilePath + " is a paranoidDirectory")
	}

	if existingFileType == typeSymlink {
		return returncodes.EIO, errors.New("existing file " + existingFilePath + " is a symlink")
	}

	targetFileType, err := getFileType(paranoidDirectory, targetParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error getting target file %s file type: %s", targetFilePath, err)
	}

	if targetFileType != typeENOENT {
		return returncodes.EEXIST, errors.New("target file " + targetFilePath + " already exists")
	}

	// getting inode and fileMode of existing file
	inodeBytes, code, err := getFileInode(existingParanoidPath)
	if code != returncodes.OK {
		return code, err
	}

	fileInfo, err := os.Stat(existingParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error stating existing file %s: %s", existingFilePath, err)
	}
	fileMode := fileInfo.Mode()

	// creating target file pointing to same inode
	err = ioutil.WriteFile(targetParanoidPath, inodeBytes, fileMode)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error writing to names file: %s", err)
	}

	// getting contents of inode
	inodePath := path.Join(paranoidDirectory, "inodes", string(inodeBytes))
	inodeContents, err := ioutil.ReadFile(inodePath)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error reading inode: %s", err)
	}

	nodeData := &inode{}
	err = json.Unmarshal(inodeContents, &nodeData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error unmarshalling inode data: %s", err)
	}

	// itterating count and saving
	nodeData.Count++
	openedFile, err := os.OpenFile(inodePath, os.O_WRONLY, 0600)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error opening file: %s", err)
	}
	defer openedFile.Close()

	err = openedFile.Truncate(0)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error truncating file: %s", err)
	}

	newJSONData, err := json.Marshal(&nodeData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error marshalling json: %s", err)
	}

	_, err = openedFile.Write(newJSONData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error writing to inode file: %s", err)
	}

	// closing file
	err = openedFile.Close()
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error closing file: %s", err)
	}

	return returncodes.OK, nil
}
