package commands

import (
	"github.com/cpssd/paranoid/pfsm/returncodes"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// SymlinkCommand creates a symbolic link
// args[0] is the init point, args[1] is the existing file, args[2] is the target file
func SymlinkCommand(args []string) {
	Log.Info("symlink command called")
	if len(args) < 3 {
		Log.Fatal("not enough arguments")
	}

	directory := args[0]
	targetFilePath := getParanoidPath(directory, args[2])

	getFileSystemLock(directory, exclusiveLock)
	defer unLockFileSystem(directory)

	// Make sure the target file not existing, if it is, quit
	if getFileType(targetFilePath) != typeENOENT {
		io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.EEXIST))
		return
	}

	uuidBytes := generateNewInode()
	uuidString := string(uuidBytes)
	Log.Verbose("symlink: uuid", uuidString)

	// Create a new file with content which is the
	// relative location of the existing file
	err := ioutil.WriteFile(targetFilePath, uuidBytes, 0600)
	checkErr("symlink", err)

	err = os.Symlink(os.DevNull, path.Join(directory, "contents", uuidString))
	checkErr("symlink", err)

	err = ioutil.WriteFile(path.Join(directory, "inodes", uuidString), []byte(args[1]), 0600)

	// Send to the server if not coming from the network
	if !Flags.Network {
		sendToServer(directory, "symlink", args[1:], nil)
	}
	io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.OK))
}