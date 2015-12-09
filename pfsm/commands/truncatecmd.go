package commands

import (
	"github.com/cpssd/paranoid/pfsm/returncodes"
	"io"
	"os"
	"path"
	"strconv"
	"syscall"
)

//TruncateCommand reduces the file given as args[1] in the paranoid-direcory args[0] to the size given in args[2]
func TruncateCommand(args []string) {
	Log.Info("truncate command given")
	if len(args) < 3 {
		Log.Fatal("Not enough arguments!")
	}
	directory := args[0]
	Log.Verbose("truncate : given directory = " + directory)

	getFileSystemLock(directory, sharedLock)
	defer unLockFileSystem(directory)

	namepath := getParanoidPath(directory, args[1])

	namepathType := getFileType(directory, namepath)
	if namepathType == typeENOENT {
		io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.ENOENT))
		return
	}

	if namepathType == typeDir {
		io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.EISDIR))
		return
	}

	if namepathType == typeSymlink {
		io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.EIO))
		return
	}

	fileNameBytes, code := getFileInode(namepath)
	if code != returncodes.OK {
		io.WriteString(os.Stdout, returncodes.GetReturnCode(code))
		return
	}
	fileName := string(fileNameBytes)

	err := syscall.Access(path.Join(directory, "contents", fileName), getAccessMode(syscall.O_WRONLY))
	if err != nil {
		io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.EACCES))
		return
	}

	getFileLock(directory, fileName, exclusiveLock)
	defer unLockFile(directory, fileName)

	Log.Verbose("truncate : truncating " + fileName)
	newsize, err := strconv.Atoi(args[2])
	if err != nil {
		Log.Fatal("error converting newsize from string to int:", err)
	}

	contentsFile, err := os.OpenFile(path.Join(directory, "contents", fileName), os.O_WRONLY, 0777)
	if err != nil {
		Log.Fatal("error opening contents file:", err)
	}

	err = contentsFile.Truncate(int64(newsize))
	if err != nil {
		Log.Fatal("error truncating file:", err)
	}

	if !Flags.Network {
		sendToServer(directory, "truncate", args[1:], nil)
	}
	io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.OK))
}
