package libpfs

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

// UnlinkCommand removes a filename link from an inode.
func UnlinkCommand(paranoidDirectory, filePath string) (returnCode returncodes.Code, returnError error) {
	log.V(1).Infof("unlink called on %s in %s", filePath, paranoidDirectory)

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

	fileParanoidPath := getParanoidPath(paranoidDirectory, filePath)
	fileType, err := getFileType(paranoidDirectory, fileParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	// checking if file exists
	if fileType == typeENOENT {
		return returncodes.ENOENT, errors.New(filePath + " does not exist")
	}

	if fileType == typeDir {
		return returncodes.EISDIR, errors.New(filePath + " is a paranoidDirectory")
	}

	// getting file inode
	inodeBytes, code, err := getFileInode(fileParanoidPath)
	if code != returncodes.OK {
		return code, err
	}

	// removing filename
	err = os.Remove(fileParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error removing file in names: %s", err)
	}

	// getting inode contents
	inodePath := path.Join(paranoidDirectory, "inodes", string(inodeBytes))
	inodeContents, err := ioutil.ReadFile(inodePath)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error reading inodes contents: %s", err)
	}

	inodeData := &inode{}
	err = json.Unmarshal(inodeContents, &inodeData)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error unmarshaling json: %s", err)
	}

	if inodeData.Count == 1 {
		// remove inode and contents
		contentsPath := path.Join(paranoidDirectory, "contents", string(inodeBytes))
		err = os.Remove(contentsPath)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error removing contents: %s", err)
		}

		err = os.Remove(inodePath)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error removing inode: %s", err)
		}
	} else {
		// subtracting one from inode count and saving
		inodeData.Count--
		err = os.Truncate(inodePath, 0)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error truncating inode path: %s", err)
		}

		dataToWrite, err := json.Marshal(inodeData)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error marshalling json: %s", err)
		}

		err = ioutil.WriteFile(inodePath, dataToWrite, 0777)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error writing to inode file: %s", err)
		}
	}

	return returncodes.OK, nil
}
