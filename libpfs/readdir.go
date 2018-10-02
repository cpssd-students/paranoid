package libpfs

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"paranoid/libpfs/returncodes"
	log "paranoid/logger"
)

//ReadDirCommand returns a list of all the files in the given paranoidDirectory
func ReadDirCommand(paranoidDirectory, dirPath string) (returnCode returncodes.Code, fileNames []string, returnError error) {
	log.V(1).Info("readdir %s from %s", dirPath, paranoidDirectory)

	err := GetFileSystemLock(paranoidDirectory, SharedLock)
	if err != nil {
		return returncodes.EUNEXPECTED, nil, err
	}

	defer func() {
		err := UnLockFileSystem(paranoidDirectory)
		if err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
			fileNames = nil
		}
	}()

	dirParanoidPath := ""

	if dirPath == "" {
		dirParanoidPath = path.Join(paranoidDirectory, "names")
	} else {
		dirParanoidPath = getParanoidPath(paranoidDirectory, dirPath)
		pathFileType, err := getFileType(paranoidDirectory, dirParanoidPath)
		if err != nil {
			return returncodes.EUNEXPECTED, nil, err
		}

		if pathFileType == typeENOENT {
			return returncodes.ENOENT, nil, fmt.Errorf("%s does not exist", dirPath)
		}

		if pathFileType == typeFile {
			return returncodes.ENOTDIR, nil, fmt.Errorf("%s is of type file", dirPath)
		}

		if pathFileType == typeSymlink {
			return returncodes.ENOTDIR, nil, fmt.Errorf("%s is of type symlink", dirPath)
		}
	}

	files, err := ioutil.ReadDir(dirParanoidPath)
	if err != nil {
		return returncodes.EUNEXPECTED, nil, fmt.Errorf("error reading paranoidDirectory %s: %s", dirPath, err)
	}

	var names []string
	for i := 0; i < len(files); i++ {
		file := files[i].Name()
		if file != "info" {
			names = append(names, file[:strings.LastIndex(file, "-")])
		}
	}
	return returncodes.OK, names, nil
}
