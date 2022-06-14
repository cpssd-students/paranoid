package libpfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

// LinkCommand creates a link of a file.
func LinkCommand(
	paranoidDirectory, existingFilePath, targetFilePath string,
) (returnCode returncodes.Code, returnError error) {
	log.Printf("linking %s with %s in %s",
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
		return returncodes.ENOENT,
			fmt.Errorf("existing file %s does not exist", existingFilePath)
	}

	if existingFileType == typeDir {
		return returncodes.EISDIR,
			fmt.Errorf("existing file %s is a paranoidDirectory", existingFilePath)
	}

	if existingFileType == typeSymlink {
		return returncodes.EIO,
			fmt.Errorf("existing file %s is a symlink", existingFilePath)
	}

	targetFileType, err := getFileType(paranoidDirectory, targetParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED,
			fmt.Errorf("error getting target file %s file type: %w", targetFilePath, err)
	}

	if targetFileType != typeENOENT {
		return returncodes.EEXIST,
			fmt.Errorf("target file %s already exists", targetFilePath)
	}

	// getting inode and fileMode of existing file
	inodeBytes, code, err := getFileInode(existingParanoidPath)
	if code != returncodes.OK {
		return code, err
	}

	fileInfo, err := os.Stat(existingParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED,
			fmt.Errorf("error stating existing file %s: %w", existingFilePath, err)
	}
	fileMode := fileInfo.Mode()

	// creating target file pointing to same inode
	err = ioutil.WriteFile(targetParanoidPath, inodeBytes, fileMode)
	if err != nil {
		return returncodes.EUNEXPECTED,
			fmt.Errorf("error writing to names file: %w", err)
	}

	// getting contents of inode
	inodePath := path.Join(paranoidDirectory, "inodes", string(inodeBytes))
	inodeContents, err := ioutil.ReadFile(inodePath)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error reading inode: %w", err)
	}

	nodeData := &inode{}
	err = json.Unmarshal(inodeContents, &nodeData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error unmarshalling inode data: %w", err)
	}

	// itterating count and saving
	nodeData.Count++
	openedFile, err := os.OpenFile(inodePath, os.O_WRONLY, 0600)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error opening file: %w", err)
	}
	defer openedFile.Close()

	err = openedFile.Truncate(0)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error truncating file: %w", err)
	}

	newJSONData, err := json.Marshal(&nodeData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error marshalling json: %w", err)
	}

	_, err = openedFile.Write(newJSONData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error writing to inode file: %w", err)
	}

	// closing file
	err = openedFile.Close()
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error closing file: %w", err)
	}

	return returncodes.OK, nil
}
