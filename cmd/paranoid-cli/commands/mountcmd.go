package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli"

	"github.com/cpssd-students/paranoid/cmd/pfsd/intercom"
	"github.com/cpssd-students/paranoid/pkg/libpfs"
	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

// Mount subcommand talks to other programs to mount the pfs filesystem.
// If the file system doesn't exist it creates it.
func Mount(c *cli.Context) {
	args := c.Args()
	if len(args) < 2 {
		cli.ShowCommandHelp(c, "mount")
		os.Exit(1)
	}
	doMount(c, args)
}

func doMount(c *cli.Context, args []string) {
	var dAddr string
	if dAddr = c.String("discovery-addr"); len(dAddr) == 0 {
		dAddr = "paranoid.discovery.razoft.net:10101"
	}
	poolPassword := c.String("pool-password")
	pfsName := args[0]
	mountPoint := args[1]

	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Error Getting Current User")
		log.Fatalf("cannot get curent user: %v", err)
	}
	pfsDir := path.Join(usr.HomeDir, ".pfs", "filesystems", pfsName)

	if _, err = os.Stat(pfsDir); os.IsNotExist(err) {
		fmt.Println("FATAL: PFS directory does not exist")
		log.Fatal("PFS directory does not exist")
	}
	if _, err = os.Stat(path.Join(pfsDir, "contents")); os.IsNotExist(err) {
		fmt.Println("FATAL: PFS directory does not include contents directory")
		log.Fatal("PFS directory does not include contents directory")
	}
	if _, err = os.Stat(path.Join(pfsDir, "meta")); os.IsNotExist(err) {
		fmt.Println("FATAL: PFS directory does not include meta directory")
		log.Fatal("PFS directory does not include meta directory")
	}
	if _, err = os.Stat(path.Join(pfsDir, "names")); os.IsNotExist(err) {
		fmt.Println("FATAL: PFS directory does not include names directory")
		log.Fatal("PFS directory does not include names directory")
	}
	if _, err = os.Stat(path.Join(pfsDir, "inodes")); os.IsNotExist(err) {
		fmt.Println("FATAL: PFS directory does not include inodes directory")
		log.Fatal("PFS directory does not include inodes directory")
	}

	if pathExists(path.Join(pfsDir, "meta/", "pfsd.pid")) {
		err = os.Remove(path.Join(pfsDir, "meta/", "pfsd.pid"))
		if err != nil {
			fmt.Println("FATAL: unable to remove daemon PID file")
			log.Fatalf("Could not remove old pfsd.pid: %v", err)
		}
	}

	attributesJSON, err := ioutil.ReadFile(path.Join(pfsDir, "meta", "attributes"))
	if err != nil {
		fmt.Println("FATAL: unable to read file system attributes")
		log.Fatalf("unable to read file system attributes: %v", err)
	}

	attributes := &fileSystemAttributes{}
	err = json.Unmarshal(attributesJSON, attributes)
	if err != nil {
		fmt.Println("FATAL: unable to read file system attributes")
		log.Fatalf("unable to read file system attributes: %v", err)
	}

	if !attributes.NetworkOff {
		_, err = net.DialTimeout("tcp", dAddr, time.Duration(5*time.Second))
		if err != nil {
			fmt.Println("FATAL: Unable to reach discovery server", err)
			log.Fatalf("Unable to reach discovery server: %v", err)
		}
	}

	poolBytes, err := ioutil.ReadFile(path.Join(pfsDir, "meta", "pool"))
	if err != nil {
		fmt.Println("FATAL: unable to read pool information:", err)
		log.Fatalf("unable to read pool information: %v", err)
	}
	pool := string(poolBytes)

	returncode, err := libpfs.MountCommand(pfsDir, dAddr, mountPoint)
	if returncode != returncodes.OK {
		fmt.Println("FATAL: Error running pfs mount command:", err)
		log.Fatalf("Error running pfs mount command: %v", err)
	}

	pfsFlags := []string{
		fmt.Sprintf("-paranoid_dir=%s", pfsDir),
		fmt.Sprintf("-mount_dir=%s", mountPoint),
		fmt.Sprintf("-discovery_addr=%s", dAddr),
		fmt.Sprintf("-discovery_pool=%s", pool),
		fmt.Sprintf("-pool_password=%s", poolPassword),
	}
	if c.GlobalBool("verbose") {
		// TODO: Pass verbosity level
		pfsFlags = append(pfsFlags, "--v=2")
	}
	if !attributes.NetworkOff {
		if iface := c.String("interface"); iface != "" {
			pfsFlags = append(pfsFlags, fmt.Sprintf("-interface=%s", iface))
		}

		certPath := path.Join(pfsDir, "meta", "cert.pem")
		keyPath := path.Join(pfsDir, "meta", "key.pem")
		if pathExists(certPath) && pathExists(keyPath) {
			log.Println("Starting PFSD in secure mode.")
			pfsFlags = append(pfsFlags, fmt.Sprintf("-cert=%s", certPath))
			pfsFlags = append(pfsFlags, fmt.Sprintf("-key=%s", keyPath))
		} else {
			if !c.Bool("noprompt") {
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Print("Starting networking in unsecure mode. Are you sure? [y/N] ")
				scanner.Scan()
				answer := strings.ToLower(scanner.Text())
				if !strings.HasPrefix(answer, "y") {
					fmt.Println("Okay. Exiting ...")
					os.Exit(1)
				}
			}
		}
	}

	cmd := exec.Command("pfsd", pfsFlags...)
	if err = cmd.Start(); err != nil {
		fmt.Println("FATAL: Error running pfsd command :", err)
		log.Fatalf("Error running pfsd command: %v", err)
	}

	// Now that we've successfully told PFSD to start, ping it until we can confirm it is up
	var ws sync.WaitGroup
	ws.Add(1)
	go func() {
		defer ws.Done()
		socketPath := path.Join(pfsDir, "meta", "intercom.sock")
		after := time.After(time.Second * 20)
		var lastConnectionError error
		for {
			select {
			case <-after:
				if lastConnectionError != nil {
					log.Println(lastConnectionError)
				}
				fmt.Printf("PFSD failed to start: received no response from PFSD. See %s for more information.\n",
					path.Join(pfsDir, "meta", "logs", "pfsd.log"))
				return
			default:
				var resp intercom.EmptyMessage
				client, err := rpc.Dial("unix", socketPath)
				if err != nil {
					time.Sleep(time.Second)
					lastConnectionError = fmt.Errorf("Could not dial pfsd unix socket: %s", err)
					continue
				}
				err = client.Call("IntercomServer.ConfirmUp", new(intercom.EmptyMessage), &resp)
				if err == nil {
					return
				}
				lastConnectionError = fmt.Errorf("Could not call pfsd confirm up over unix socket: %s", err)
				time.Sleep(time.Second)
			}
		}
	}()
	ws.Wait()
}
