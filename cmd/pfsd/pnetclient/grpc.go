package pnetclient

import (
	"context"
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
)

// Dial a node and return a connection if successful
func Dial(node globals.Node) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials = insecure.NewCredentials()
	if globals.TLSEnabled {
		creds = credentials.NewTLS(&tls.Config{
			ServerName:         node.CommonName,
			InsecureSkipVerify: globals.TLSSkipVerify,
		})
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return grpc.DialContext(ctx, node.String(), grpc.WithTransportCredentials(creds))
}
