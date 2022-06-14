package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
	log "github.com/cpssd-students/paranoid/pkg/logger"
	"github.com/cpssd-students/paranoid/pkg/raft"
	pb "github.com/cpssd-students/paranoid/proto/raft"

	progress "github.com/cheggaaa/pb"
	"github.com/urfave/cli"
)

// Buildfs creates a new filesystem from a set of activity logs specified
func Buildfs(c *cli.Context) {
	args := c.Args()
	if len(args) != 2 {
		cli.ShowCommandHelp(c, "buildfs")
		os.Exit(1)
	}

	pfsName := args[0]
	logDir := args[1]

	if fileSystemExists(pfsName) {
		fmt.Println(pfsName, "already exists. please chose a different name")
		os.Exit(1)
	}

	logs, err := ioutil.ReadDir(logDir)
	if err != nil {
		fmt.Println("Couldn't read log-directory")
		log.Fatal(err)
	}

	if len(logs) < 1 {
		fmt.Println("log directory is empty")
		log.Fatal("log directory is empty")
	}

	doInit(pfsName, c.String("pool"), c.String("cert"),
		c.String("key"), c.Bool("unsecure"), c.Bool("unencrypted"), c.Bool("networkoff"))

	u, err := user.Current()
	if err != nil {
		fmt.Println("Couldn't get current user home directory")
		log.Fatalf("could not get current user home directory: %v", err)
	}

	pfsPath := path.Join(u.HomeDir, ".pfs", "filesystems", pfsName)
	err = os.MkdirAll(path.Join(pfsPath, "meta", "raft", "raft_logs"), 0700)
	if err != nil {
		cleanupPFS(pfsPath)
		log.Fatalf("could not make a raft log directory: %v", err)
	}

	bar := progress.StartNew(len(logs))
	for _, lg := range logs {
		logEntry, err := fileToProto(lg, logDir)
		if err != nil {
			cleanupPFS(pfsPath)
			fmt.Println("broken file in log directory: ", lg)
			log.Fatalf("broken file in log directory %s: %v", lg, err)
		}

		if logEntry.Entry.Type == pb.Entry_StateMachineCommand {
			er := raft.PerformLibPfsCommand(pfsPath, logEntry.Entry.Command)
			if er.Code == returncodes.EUNEXPECTED {
				fmt.Println("pfsLib failed on command: ", logEntry.Entry.Command)
				log.Fatalf("libpfs failed on command %s: %v",
					logEntry.Entry.Command.String(), er.Err)
			}
		}
		bar.Increment()
	}

	bar.Finish()
	fmt.Printf("Done.\nYou may now mount your new filesystem: %s", pfsName)
}
