package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"

	"github.com/urfave/cli"
)

//Unmount unmounts a paranoid file system
func Unmount(c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		_ = cli.ShowCommandHelp(c, "unmount")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Error Getting Current User")
		log.Fatalf("cannot get curent user: %v", err)
	}

	pidPath := path.Join(usr.HomeDir, ".pfs", "filesystems", args[0], "meta", "pfsd.pid")
	if _, err = os.Stat(pidPath); err != nil {
		fmt.Println("FATAL: Could not read pid file")
		log.Fatal("Could not read pid file:", err)
	}

	pidByte, err := ioutil.ReadFile(pidPath)
	if err != nil {
		fmt.Println("FATAL: Can't read pid file")
		log.Fatalf("unable to read pid file: %v", err)
	}

	pid, _ := strconv.Atoi(string(pidByte))
	if err = syscall.Kill(pid, syscall.SIGTERM); err != nil {
		fmt.Println("FATAL: Can not kill PFSD")
		log.Fatalf("Can not kill PFSD: %v", err)
	}
}
