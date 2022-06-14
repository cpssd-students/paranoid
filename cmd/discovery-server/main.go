package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/cpssd-students/paranoid/cmd/discovery-server/dnetserver"

	pb "github.com/cpssd-students/paranoid/proto/paranoid/discoverynetwork/v1"
)

const (
	// DiscoveryStateDir defines where the state should be kept
	DiscoveryStateDir string = "discovery_state"
	// TempStateDir defines where the state should be temporarely kept
	TempStateDir string = ".tmp_state"
)

var (
	port          = flag.Int("port", 10101, "port to listen on")
	renewInterval = flag.Int("renew-interval", 5*60*1000, "time after which membership expires, in ms")
	certFile      = flag.String("cert", "", "TLS certificate file - if empty connection will be unencrypted")
	keyFile       = flag.String("key", "", "TLS key file - if empty connection will be unencrypted")
	loadState     = flag.Bool("state", true, "Load the Nodes from the statefile")
)

func createRPCServer() *grpc.Server {
	var opts []grpc.ServerOption
	if *certFile != "" && *keyFile != "" {
		log.Println("Starting discovery server with TLS.")
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate TLS credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	} else {
		log.Println("Starting discovery server without TLS.")
	}
	return grpc.NewServer(opts...)
}

func main() {
	flag.Parse()
	dnetserver.Pools = make(map[string]*dnetserver.Pool)

	analyseWorkspace()

	renewDuration, err := time.ParseDuration(strconv.Itoa(*renewInterval) + "ms")
	if err != nil {
		log.Printf("Failed parsing renew interval: %v", err)
	}

	dnetserver.RenewInterval = renewDuration

	log.Println("Starting Paranoid Discovery Server")

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v.", *port, err)
	}
	log.Printf("Listening on %s", lis.Addr().String())

	if *loadState {
		dnetserver.LoadState()
	}
	srv := createRPCServer()
	pb.RegisterDiscoveryNetworkServiceServer(srv, &dnetserver.DiscoveryServer{})

	log.Println("gRPC server created")
	_ = srv.Serve(lis)
}

// analyseWorkspace analyses the state of the workspace directory for the server,
// if the workspace directory doesnt exist it will be recreated along with needed
// sub-directories.
func analyseWorkspace() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Couldn't identify user:", err)
	}

	// checking ~/.pfs
	pfsDirPath := path.Join(usr.HomeDir, ".pfs")
	checkDir(pfsDirPath)

	// checking ~/.pfs/discovery_meta
	metaDirPath := path.Join(pfsDirPath, "discovery_meta")
	checkDir(metaDirPath)

	dnetserver.StateDirectoryPath = path.Join(metaDirPath, DiscoveryStateDir)
	dnetserver.TempDirectoryPath = path.Join(metaDirPath, TempStateDir)
	checkDir(dnetserver.StateDirectoryPath)
	checkDir(dnetserver.TempDirectoryPath)
}

// checkDir checks a directory and creates it if needed
func checkDir(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Creating %s", dir)

			if err := os.Mkdir(dir, 0700); err != nil {
				log.Fatalf("Failed to create %s: %v", dir, err)
			}
		} else {
			log.Fatal("Couldn't stat:", dir, "err:", err)
		}
	} else {

		if err := syscall.Access(dir, syscall.O_RDWR); err != nil {
			log.Fatalf("Don't have read & write access to %s: %v", dir, err)
		}
	}
}
