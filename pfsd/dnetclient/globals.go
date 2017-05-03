package dnetclient

import (
	"crypto/tls"
	"time"

	"github.com/pp2p/paranoid/logger"
	"github.com/pp2p/paranoid/pfsd/globals"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const peerPingTimeOut time.Duration = time.Minute * 3
const peerPingInterval time.Duration = time.Minute

var (
	discoveryCommonName string

	Log *logger.ParanoidLogger
)

func dialDiscovery() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(2*time.Second))
	if globals.TLSEnabled {
		creds := credentials.NewTLS(&tls.Config{
			ServerName:         discoveryCommonName,
			InsecureSkipVerify: globals.TLSSkipVerify,
		})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	return grpc.Dial(globals.DiscoveryAddr, opts...)
}
