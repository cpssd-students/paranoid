package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"

	"github.com/urfave/cli"

	log "github.com/cpssd-students/paranoid/pkg/logger"
)

// Restart subcommand
func Restart(c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		_ = cli.ShowCommandHelp(c, "restart")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Could not get current user")
		log.Fatalf("could not get current user: %v", err)
	}

	pidPath := path.Join(usr.HomeDir, ".pfs", "filesystems", args[0], "meta", "pfsd.pid")
	_, err = os.Stat(pidPath)
	if err != nil {
		fmt.Println("FATAL: Could not access PID file")
		log.Fatalf("Could not access PID file: %v", err)
	}

	pidByte, err := ioutil.ReadFile(pidPath)
	if err != nil {
		fmt.Println("FATAL: Can't read PID file")
		log.Fatalf("cannot read PID file: %v", err)
	}
	pid, _ := strconv.Atoi(string(pidByte))
	if err = syscall.Kill(pid, syscall.SIGHUP); err != nil {
		fmt.Println("FATAL: Can not restart PFSD")
		log.Fatalf("Can not restart PFSD: %v", err)
	}
}
