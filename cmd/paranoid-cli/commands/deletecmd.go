package commands

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/urfave/cli"
)

//Delete deletes a paranoid file system
func Delete(c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		_ = cli.ShowCommandHelp(c, "delete")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Error Getting Current User")
		log.Fatalf("cannot get curent user: %v", err)
	}

	pfspath, err := filepath.Abs(path.Join(usr.HomeDir, ".pfs", "filesystems", args[0]))
	if err != nil {
		fmt.Println("FATAL: Paranoid file system name is incorrectly formatted")
		log.Fatalf("given pfs-name is in incorrect format: %v", err)
	}
	if path.Base(pfspath) != args[0] {
		fmt.Println("FATAL: Paranoid file system name is incorrectly formatted")
		log.Fatalf("given pfs-name is in incorrect format")
	}

	err = os.RemoveAll(pfspath)
	if err != nil {
		fmt.Println("FATAL: Could not delete given paranoid file system.")
		log.Fatalf("could not delete given paranoid file system: %v", err)
	}
}
