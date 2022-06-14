package pfi

import (
	"log"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// StartPfi with given verbosity level
func StartPfi() {
	// Create a logger
	var err error

	if globals.RaftNetworkServer == nil {
		SendOverNetwork = false
	} else {
		SendOverNetwork = true
	}

	// setting up with fuse
	opts := pathfs.PathNodeFsOptions{}
	opts.ClientInodes = true
	nfs := pathfs.NewPathNodeFs(&ParanoidFileSystem{
		FileSystem: pathfs.NewDefaultFileSystem(),
	}, &opts)
	server, _, err := nodefs.MountRoot(globals.MountPoint, nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v", err)
	}

	globals.Wait.Add(1)
	go func() {
		defer globals.Wait.Done()
		globals.Wait.Add(1)
		go func() {
			defer globals.Wait.Done()
			server.Serve()
		}()

		if ok := <-globals.Quit; !ok {
			log.Print("Attempting to unmount pfi")
			err = server.Unmount()
			if err != nil {
				log.Fatalf("Error unmounting: %v", err)
			}
			log.Print("pfi unmounted successfully")
			return
		}
	}()
}
