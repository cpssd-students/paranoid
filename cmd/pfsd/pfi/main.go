package pfi

import (
	"path"

	"paranoid/pkg/logger"
	"paranoid/cmd/pfsd/globals"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// StartPfi with given verbosity level
func StartPfi(logVerbose bool) {
	// Create a logger
	var err error
	Log = logger.New("pfi", "pfsd", path.Join(globals.ParanoidDir, "meta", "logs"))
	Log.SetOutput(logger.STDERR | logger.LOGFILE)

	if globals.RaftNetworkServer == nil {
		SendOverNetwork = false
	} else {
		SendOverNetwork = true
	}

	if logVerbose {
		Log.SetLogLevel(logger.VERBOSE)
	}

	// setting up with fuse
	opts := pathfs.PathNodeFsOptions{}
	opts.ClientInodes = true
	nfs := pathfs.NewPathNodeFs(&ParanoidFileSystem{
		FileSystem: pathfs.NewDefaultFileSystem(),
	}, &opts)
	server, _, err := nodefs.MountRoot(globals.MountPoint, nfs.Root(), nil)
	if err != nil {
		Log.Fatalf("Mount fail: %v\n", err)
	}

	globals.Wait.Add(1)
	go func() {
		defer globals.Wait.Done()
		globals.Wait.Add(1)
		go func() {
			defer globals.Wait.Done()
			server.Serve()
		}()

		select {
		case _, ok := <-globals.Quit:
			if !ok {
				Log.Info("Attempting to unmount pfi")
				err = server.Unmount()
				if err != nil {
					Log.Fatal("Error unmounting : ", err)
				}
				Log.Info("pfi unmounted successfully")
				return
			}
		}
	}()
}
