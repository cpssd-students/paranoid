package pnetclient

import (
	"crypto/tls"
	"github.com/pp2p/paranoid/logger"
	"github.com/pp2p/paranoid/pfsd/globals"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

var Log *logger.ParanoidLogger

func Dial(node globals.Node) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(5*time.Second))
	if globals.TLSEnabled {
		creds := credentials.NewTLS(&tls.Config{
			ServerName:         node.CommonName,
			InsecureSkipVerify: globals.TLSSkipVerify,
		})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(node.String(), opts...)
	return conn, err
}
