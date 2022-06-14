package intercom

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/cpssd-students/paranoid/pkg/raft"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
)

// Message constants
const (
	StatusNetworkOff string = "Networking disabled"
)

// Server listens on a Unix socket to respond to questions asked by
// external applications.
type Server struct{}

// EmptyMessage with no data
type EmptyMessage struct{}

// StatusResponse with data about the status of pfsd
type StatusResponse struct {
	Uptime    time.Duration
	Status    string
	TLSActive bool
	Port      int
}

// ListNodesResponse with all Nodes
type ListNodesResponse struct {
	Nodes []raft.Node
}

// ConfirmUp is simple method that paranoid-cli uses to ping PFSD
func (s *Server) ConfirmUp(req *EmptyMessage, resp *EmptyMessage) error {
	return nil
}

// Status provides health data for the current node.
func (s *Server) Status(req *EmptyMessage, resp *StatusResponse) error {
	if globals.NetworkOff {
		resp.Uptime = time.Since(globals.BootTime)
		resp.Status = StatusNetworkOff
		return nil
	}

	thisport, err := strconv.Atoi(globals.ThisNode.Port)
	if err != nil {
		log.Printf("Could not convert globals.ThisNode.Port to int.")
		return fmt.Errorf("failed converting globals.ThisNode.Port to int: %s", err)
	}

	resp.Uptime = time.Since(globals.BootTime)
	if globals.RaftNetworkServer != nil {
		switch globals.RaftNetworkServer.State.GetCurrentState() {
		case raft.FOLLOWER:
			resp.Status = "Follower"
		case raft.CANDIDATE:
			resp.Status = "Candidate"
		case raft.LEADER:
			resp.Status = "Leader"
		case raft.INACTIVE:
			resp.Status = "Raft Inactive"
		}
	} else {
		resp.Status = "Networking Disabled"
	}
	resp.TLSActive = globals.TLSEnabled
	resp.Port = thisport
	return nil
}

// ListNodes that pfsd is connected to
func (s *Server) ListNodes(req *EmptyMessage, resp *ListNodesResponse) error {
	if globals.RaftNetworkServer == nil {
		return fmt.Errorf("Networking Disabled")
	}
	resp.Nodes = globals.RaftNetworkServer.State.Configuration.GetPeersList()
	return nil
}

// RunServer starts the intercom server on a socket stored in the specified
// meta directory
func RunServer(metaDir string) {
	socketPath := path.Join(metaDir, "intercom.sock")
	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to remove %s: %s\n", socketPath, err)
	}
	server := new(Server)
	_ = rpc.Register(server)
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %s\n", socketPath, err)
	}
	globals.Wait.Add(1)
	go func() {
		defer globals.Wait.Done()
		log.Printf("Internal communication server listening on %s", socketPath)
		globals.Wait.Add(1)
		go func() {
			rpc.Accept(lis)
			defer globals.Wait.Done()
		}()

		if ok := <-globals.Quit; !ok {
			log.Print("Stopping internal communication server")
			err := lis.Close()
			if err != nil {
				log.Printf("Could not shut down internal communication server: %v", err)
			} else {
				log.Print("Internal communication server stopped.")
			}
		}
	}()
	globals.BootTime = time.Now()
	log.Print(globals.BootTime)
}
