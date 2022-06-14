package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/kardianos/osext"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/cmd/pfsd/upnp"
)

func stopAllServices() {
	globals.ShuttingDown = true
	if globals.UPnPEnabled {
		err := upnp.ClearPortMapping(globals.ThisNode.Port)
		if err != nil {
			log.Printf("Could not clear port mapping: %v", err)
		}
	}
	close(globals.Quit) // Sends stop signal to all goroutines

	// Save all KeyPieces to disk, to ensure we haven't missed any so far.
	_ = globals.HeldKeyPieces.SaveToDisk()

	if !globals.NetworkOff {
		close(globals.RaftNetworkServer.Quit)
		srv.Stop()
		globals.RaftNetworkServer.Wait.Wait()
	}

	globals.Wait.Wait()
}

// HandleSignals listens for SIGTERM and SIGHUP, and dispatches to handler
// functions when a signal is received.
func HandleSignals() {
	incoming := make(chan os.Signal, 1)
	signal.Notify(incoming, syscall.SIGHUP, syscall.SIGTERM)
	sig := <-incoming
	switch sig {
	case syscall.SIGHUP:
		handleSIGHUP()
	case syscall.SIGTERM:
		handleSIGTERM()
	}
}

func handleSIGHUP() {
	log.Print("SIGHUP received. Restarting.")
	stopAllServices()
	log.Print("All services stopped. Forking process.")
	execSpec := &syscall.ProcAttr{
		Env: os.Environ(),
	}
	pathToSelf, err := osext.Executable()
	if err != nil {
		log.Printf("Could not get path to self: %v", err)
		pathToSelf = os.Args[0]
	}
	fork, err := syscall.ForkExec(pathToSelf, os.Args, execSpec)
	if err != nil {
		log.Printf("Could not fork child PFSD instance: %v", err)
	} else {
		log.Printf("Forked successfully. New PID: %v", fork)
	}
}

func handleSIGTERM() {
	log.Print("SIGTERM received. Exiting.")
	stopAllServices()
	err := os.Remove(path.Join(globals.ParanoidDir, "meta", "pfsd.pid"))
	if err != nil {
		log.Printf("Can't remove PID file: %v", err)
	}
	log.Print("All services stopped. Have a nice day.")
}
