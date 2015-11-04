package commands

import (
	"github.com/cpssd/paranoid/pfsm/returncodes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
)

//MountCommand is used to notify a pfs directory it has been mounted.
//Stores the ip given as args[1] and the port given as args[2] in files in the meta directory.
func MountCommand(args []string) {
	verboseLog("mount command called")
	if len(args) < 3 {
		log.Fatalln("Not enough arguments!")
	}
	directory := args[0]
	verboseLog("mount : given directory = " + directory)
	err := ioutil.WriteFile(path.Join(directory, "meta", "ip"), []byte(args[1]), 0777)
	checkErr("mount", err)
	err = ioutil.WriteFile(path.Join(directory, "meta", "port"), []byte(args[2]), 0777)
	checkErr("mount", err)
	io.WriteString(os.Stdout, returncodes.GetReturnCode(returncodes.OK))
}