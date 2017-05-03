// Package pnetserver implements the ParanoidNetwork gRPC server.
// globals.go contains data used by each gRPC handler in pnetserver.
package pnetserver

import (
	"github.com/pp2p/paranoid/logger"
)

type ParanoidServer struct{}

var Log *logger.ParanoidLogger
