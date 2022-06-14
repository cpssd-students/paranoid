package dnetclient

import (
	"context"
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
)

const peerPingTimeOut time.Duration = time.Minute * 3
const peerPingInterval time.Duration = time.Minute

var (
	discoveryCommonName string
)

func dialDiscovery() (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials = insecure.NewCredentials()
	if globals.TLSEnabled {
		creds = credentials.NewTLS(&tls.Config{
			ServerName:         discoveryCommonName,
			InsecureSkipVerify: globals.TLSSkipVerify,
		})
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return grpc.DialContext(ctx, globals.DiscoveryAddr, grpc.WithTransportCredentials(creds))
}
