package commands

import (
	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/cpssd-students/paranoid/cmd/pfsd/intercom"
	log "github.com/cpssd-students/paranoid/pkg/logger"
)

// Status displays statistics for the specified PFSD instances.
func Status(c *cli.Context) {
	args := c.Args()
	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Couldn't get current user")
		log.Fatalf("cannot get current user: %v", err)
	}

	// By default, print the status of each running instance.
	if !args.Present() {
		dirs, err := ioutil.ReadDir(path.Join(usr.HomeDir, ".pfs", "filesystems"))
		if err != nil {
			fmt.Println("FATAL: Could not get absolute path to paranoid filesystem.")
			log.Fatalf("could not get absolute path to paranoid filesystem: %v", err)
		}
		for _, dir := range dirs {
			dirPath := path.Join(usr.HomeDir, ".pfs", "filesystems", dir.Name())
			if _, err := os.Stat(path.Join(dirPath, "meta", "pfsd.pid")); err == nil {
				getStatus(dirPath)
			}
		}
	} else {
		for _, dir := range args {
			getStatus(path.Join(usr.HomeDir, ".pfs", "filesystems", dir))
		}
	}
}

func getStatus(pfsDir string) {
	// We check this on the off chance they haven't initialised a single PFS yet.
	if _, err := os.Stat(pfsDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s does not exist. Please call 'paranoid-cli init' before running this command.", pfsDir)
			log.Fatalf("PFS directory %s does not exist", pfsDir)
		} else {
			fmt.Printf("Could not stat %s: %s.", pfsDir, err)
			log.Fatalf("Could not stat PFS directory %s: %v", pfsDir, err)
		}
	}

	socketPath := path.Join(pfsDir, "meta", "intercom.sock")
	logPath := path.Join(pfsDir, "meta", "logs", "pfsd.log")
	var resp intercom.StatusResponse
	client, err := rpc.Dial("unix", socketPath)
	if err != nil {
		fmt.Printf("Could not connect to PFSD %s. Is it running? See %s for more information.\n", filepath.Base(pfsDir), logPath)
		log.Warnf("Could not connect to PFSD %s at %s: %s", filepath.Base(pfsDir), socketPath, err)
		return
	}
	err = client.Call("IntercomServer.Status", new(intercom.EmptyMessage), &resp)
	if err != nil {
		fmt.Printf("Error getting status for %s. See %s for more information.\n", filepath.Base(pfsDir), logPath)
		log.Warnf("PFSD at %s returned error: %s", filepath.Base(pfsDir), err)
		return
	}
	printStatusInfo(filepath.Base(pfsDir), resp)
}

func printStatusInfo(pfsName string, info intercom.StatusResponse) {
	fmt.Printf("\nFilesystem Name:\t%s\n", pfsName)
	fmt.Printf("Uptime:\t\t\t%s\n", info.Uptime.String())
	fmt.Printf("Raft Status:\t\t%s\n", info.Status)
	if info.Status != intercom.StatusNetworkOff {
		fmt.Printf("TLS Enabled:\t\t%t\n", info.TLSActive)
		fmt.Printf("Port:\t\t\t%d\n", info.Port)
	}
}
