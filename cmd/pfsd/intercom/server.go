package intercom

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path"
	"strconv"
	"time"

	"paranoid/cmd/pfsd/globals"
	"paranoid/pkg/raft"
)

// Message constants
const (
	StatusNetworkOff string = "Networking disabled"
)

// IntercomServer listens on a Unix socket to respond to questions asked by
// external applications.
type IntercomServer struct{}

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
func (s *IntercomServer) ConfirmUp(req *EmptyMessage, resp *EmptyMessage) error {
	return nil
}

// Status provides health data for the current node.
func (s *IntercomServer) Status(req *EmptyMessage, resp *StatusResponse) error {
	if globals.NetworkOff {
		resp.Uptime = time.Since(globals.BootTime)
		resp.Status = StatusNetworkOff
		return nil
	}

	thisport, err := strconv.Atoi(globals.ThisNode.Port)
	if err != nil {
		Log.Error("Could not convert globals.ThisNode.Port to int.")
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
func (s *IntercomServer) ListNodes(req *EmptyMessage, resp *ListNodesResponse) error {
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
		Log.Fatalf("Failed to remove %s: %s\n", socketPath, err)
	}
	server := new(IntercomServer)
	rpc.Register(server)
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		Log.Fatalf("Failed to listen on %s: %s\n", socketPath, err)
	}
	globals.Wait.Add(1)
	go func() {
		defer globals.Wait.Done()
		Log.Info("Internal communication server listening on", socketPath)
		globals.Wait.Add(1)
		go func() {
			rpc.Accept(lis)
			defer globals.Wait.Done()
		}()

		select {
		case _, ok := <-globals.Quit:
			if !ok {
				Log.Info("Stopping internal communication server")
				err := lis.Close()
				if err != nil {
					Log.Warn("Could not shut down internal communication server:", err)
				} else {
					Log.Info("Internal communication server stopped.")
				}
			}
		}
	}()
	globals.BootTime = time.Now()
	Log.Info(globals.BootTime)
}
