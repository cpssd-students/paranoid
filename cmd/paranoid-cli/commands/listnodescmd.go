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

	"paranoid/cmd/pfsd/intercom"
	log "paranoid/pkg/logger"
	"paranoid/pkg/raft"
)

// ListNodes subcommand shows the statistics about nodes
func ListNodes(c *cli.Context) {
	args := c.Args()
	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Unable to get information on current user:", err)
		log.Fatalf("could not get user information: %v", err)
	}

	// By default, list the nodes connected to each running instance.
	if !args.Present() {
		dirs, err := ioutil.ReadDir(path.Join(usr.HomeDir, ".pfs", "filesystems"))
		if err != nil {
			fmt.Printf("FATAL: Unable to get list of paranoid file systems. Does %s exist?", path.Join(usr.HomeDir, ".pfs"))
			log.Fatalf("could not get list of paranoid file systems: %v", err)
		}
		for _, dir := range dirs {
			dirPath := path.Join(usr.HomeDir, ".pfs", "filesystems", dir.Name())
			if _, err := os.Stat(path.Join(dirPath, "meta", "pfsd.pid")); err == nil {
				getNodes(dirPath)
			}
		}
	} else {
		for _, dir := range args {
			getNodes(path.Join(usr.HomeDir, ".pfs", "filesystems", dir))
		}
	}
}

func getNodes(pfsDir string) {
	// We check this on the off chance they haven't initialised a single PFS yet.
	if _, err := os.Stat(pfsDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s does not exist. Please call 'paranoid-cli init' before running this command.", pfsDir)
			log.Fatal("PFS directory does not exist.")
		} else {
			fmt.Printf("Could not stat %s. Error returned: %s.", pfsDir, err)
			log.Fatalf("could not stat PFS directory: %v", err)
		}
	}

	socketPath := path.Join(pfsDir, "meta", "intercom.sock")
	logPath := path.Join(pfsDir, "meta", "logs", "pfsd.log")
	var resp intercom.ListNodesResponse
	client, err := rpc.Dial("unix", socketPath)
	if err != nil {
		fmt.Printf("Could not connect to PFSD %s. Is it running? See %s for more information.\n", filepath.Base(pfsDir), logPath)
		log.Warnf("Could not connect to PFSD %s at %s: %v", filepath.Base(pfsDir), socketPath, err)
		return
	}
	err = client.Call("IntercomServer.ListNodes", new(intercom.EmptyMessage), &resp)
	if err != nil {
		if err.Error() == "Networking Disabled" {
			fmt.Println("Networking is disabled")
		} else {
			fmt.Printf("Error listing nodes connected to %s. See %s for more information.\n", filepath.Base(pfsDir), logPath)
			log.Warnf("PFSD at %s returned error: %v", filepath.Base(pfsDir), err)
		}
		return
	}
	printAllNodes(filepath.Base(pfsDir), resp)
}

func printAllNodes(pfsName string, info intercom.ListNodesResponse) {
	fmt.Printf("\n----- Nodes Connected to %s -----\n", pfsName)
	for _, node := range info.Nodes {
		printSingleNode(node)
	}
}

func printSingleNode(node raft.Node) {
	fmt.Printf("IP: \t\t%s\n", node.IP)
	fmt.Printf("Port: \t\t%s\n", node.Port)
	fmt.Printf("UUID: \t\t%s\n", node.NodeID)
}
