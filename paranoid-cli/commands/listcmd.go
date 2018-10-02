package commands

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path"

	"github.com/urfave/cli"

	log "paranoid/logger"
)

//List lists all paranoid file systems
func List(c *cli.Context) {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Error Getting Current User")
		log.Fatalf("cannot get curent user: %v", err)
	}

	files, err := ioutil.ReadDir(path.Join(usr.HomeDir, ".pfs", "filesystems"))
	if err != nil {
		fmt.Println("FATAL: Could not read the list of paranoid file systems.")
		log.Fatalf("Could not get list of paranoid file systems: %v", err)
	}
	for _, file := range files {
		fmt.Println(file.Name())
	}
}
