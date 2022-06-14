// Package dnetserver implements the DiscoveryNetwork gRPC server.
// globals.go contains data used by each gRPC handler in dnetserver.
package dnetserver

import (
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/cpssd-students/paranoid/proto/paranoid/discoverynetwork/v1"
)

// DiscoveryServer struct
type DiscoveryServer struct {
	pb.UnimplementedDiscoveryNetworkServer
}

// PoolInfo struct to hold the pool data
type PoolInfo struct {
	PasswordSalt []byte `json:"passwordsalt"`
	PasswordHash []byte `json:"passwordhash"`
	Nodes        map[string]*pb.Node
}

// Pool is the safe wrapper around PoolInfo with a mutex.
type Pool struct {
	PoolLock sync.Mutex
	Info     PoolInfo
}

// PoolLock is locked when acessing the pool map
var PoolLock sync.RWMutex

// Pools map
var Pools map[string]*Pool

// RenewInterval global containing the time after which the nodes will be marked as inactive
var RenewInterval time.Duration

// StateDirectoryPath is the path to the directory in which the discovery server stores its state
var StateDirectoryPath string

// TempDirectoryPath is the path to the directory where temporary state files are stored
var TempDirectoryPath string

func checkPoolPassword(pool, password string, node *pb.Node) error {
	if _, ok := Pools[pool]; ok {
		if password == "" {
			if len(Pools[pool].Info.PasswordHash) != 0 {
				log.Printf("Join: node %s attempted join password protected pool "+
					"without a giving a password",
					node.Uuid)
				returnError := status.Errorf(codes.Internal,
					"pool %s is password protected",
					pool,
				)
				return returnError
			}
		} else {
			if err := bcrypt.CompareHashAndPassword(
				Pools[pool].Info.PasswordHash,
				append(Pools[pool].Info.PasswordSalt, []byte(password)...),
			); err != nil {
				log.Printf("Join: node %s attempted join password protected pool "+
					"with incorrect password: %v",
					node.Uuid,
					err,
				)
				return status.Errorf(codes.Internal, "given password incorrect: %v", err)
			}
		}
	}
	return nil
}
