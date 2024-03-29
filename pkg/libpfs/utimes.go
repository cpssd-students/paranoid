package libpfs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

//UtimesCommand updates the acess time and modified time of a file
func UtimesCommand(
	paranoidDirectory, filePath string, atime, mtime *time.Time,
) (returnCode returncodes.Code, returnError error) {
	log.Printf("utimes called on %s in %s", filePath, paranoidDirectory)

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

	namepath := getParanoidPath(paranoidDirectory, filePath)

	fileType, err := getFileType(paranoidDirectory, namepath)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	if fileType == typeENOENT {
		return returncodes.ENOENT, errors.New(filePath + " does not exist")
	}

	fileInodeBytes, code, err := getFileInode(namepath)
	if code != returncodes.OK {
		return code, err
	}
	inodeName := string(fileInodeBytes)

	err = getFileLock(paranoidDirectory, inodeName, ExclusiveLock)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	defer func() {
		err := unLockFile(paranoidDirectory, inodeName)
		if err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
		}
	}()

	file, err := os.Open(path.Join(paranoidDirectory, "contents", inodeName))
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error opening contents file: %s", err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error stating file: %s", err)
	}

	stat := fi.Sys().(*syscall.Stat_t)
	oldatime := lastAccess(stat)
	oldmtime := fi.ModTime()
	if atime == nil && mtime == nil {
		return returncodes.EUNEXPECTED, errors.New("no times to update")
	}

	if atime == nil {
		err = os.Chtimes(path.Join(paranoidDirectory, "contents", inodeName), oldatime, *mtime)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error changing times: %s", err)
		}
	} else if mtime == nil {
		err = os.Chtimes(path.Join(paranoidDirectory, "contents", inodeName), *atime, oldmtime)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error changing times: %s", err)
		}
	} else {
		err = os.Chtimes(path.Join(paranoidDirectory, "contents", inodeName), *atime, *mtime)
		if err != nil {
			return returncodes.EUNEXPECTED, fmt.Errorf("error changing times: %s", err)
		}
	}

	return returncodes.OK, nil
}
