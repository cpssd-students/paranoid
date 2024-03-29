package dnetclient

import (
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/cpssd-students/paranoid/cmd/pfsd/pnetclient"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
)

// SetDiscovery sets up the discovery server to use
func SetDiscovery(addr string) {
	globals.DiscoveryAddr = addr

	host := strings.Split(addr, ":")[0]
	if globals.TLSEnabled && !globals.TLSSkipVerify {
		// If host is an IP, we need to get the hostname via DNS
		ip := net.ParseIP(host)
		var dnsAddr string
		var err error
		if ip != nil {
			var dnsAddrs []string
			dnsAddrs, err = net.LookupAddr(ip.String())
			dnsAddr = dnsAddrs[0]
		} else {
			dnsAddr, err = net.LookupCNAME(host)
		}
		if err != nil {
			log.Printf("Could not complete DNS lookup: %v", err)
		}
		if dnsAddr == "" { // If no DNS entries exist
			log.Fatalf("Can not find DNS entry for discovery server %s", host)
		}
		discoveryCommonName = dnsAddr
	}
}

// JoinDiscovery connects to the selected discovery server
func JoinDiscovery(pool, password string) {
	if err := Join(pool, password); err != nil {
		if err = retryJoin(pool, password); err != nil {
			log.Fatalf("Failure dialing discovery server after multiple attempts, Giving up: %v",
				err)
		}
	}
	globals.Wait.Add(1)
	go pingPeers()
}

// Periodically pings all known nodes on the network. Lives here and not
// in pnetclient since pnetclient is stateless and this function is more
// relevant to discovery.
func pingPeers() {
	defer globals.Wait.Done()
	// Ping as soon as this node joins
	pnetclient.Ping()
	timer := time.NewTimer(peerPingInterval)
	defer timer.Stop()
	for {
		select {
		case _, ok := <-globals.Quit:
			if !ok {
				return
			}
		case <-timer.C:
			pnetclient.Ping()
			timer.Reset(peerPingInterval)
		}
	}
}

//JoinCluster sends a request to all peers to request to be added to the cluster
func JoinCluster(password string) error {
	timer := time.NewTimer(peerPingTimeOut)
	defer timer.Stop()
	for {
		select {
		case _, ok := <-globals.Quit:
			if !ok {
				return nil
			}
		case <-timer.C:
			return errors.New("Failed to join raft cluster")
		default:
			err := pnetclient.JoinCluster(password)
			if err == nil {
				return nil
			}
		}
	}
}

func retryJoin(pool, password string) error {
	var err error
	for i := 0; i < 10; i++ {
		err = Join(pool, password)
		if err == nil {
			break
		}
	}
	return err
}
