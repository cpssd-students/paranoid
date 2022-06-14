// Package pnetserver implements the ParanoidNetwork gRPC server.
// globals.go contains data used by each gRPC handler in pnetserver.
package pnetserver

import (
	"paranoid/pkg/logger"
)

// ParanoidServer implements the paranoidnetwork gRPC server
type ParanoidServer struct{}

// Log used by pnetserver
var Log *logger.ParanoidLogger
