package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/urfave/cli"

	log "github.com/cpssd-students/paranoid/pkg/logger"
)

//AutoMount mounts a file system with the last used settings.
func AutoMount(c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		cli.ShowCommandHelp(c, "automount")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Error Getting Current User")
		log.Fatalf("cannot get curent user: %v", err)
	}
	pfsMeta := path.Join(usr.HomeDir, ".pfs", "filesystems", args[0], "meta")

	mountpoint, err := ioutil.ReadFile(path.Join(pfsMeta, "mountpoint"))
	if err != nil {
		fmt.Println("FATAL: PFSD Couldnt find FS mountpoint", err)
		log.Fatalf("could not get mountpoint %v", err)
	}

	mountArgs := []string{args[0], string(mountpoint)}
	doMount(c, mountArgs)
}
