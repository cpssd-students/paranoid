package upnp

import (
	"errors"
	"flag"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"github.com/huin/goupnp/dcps/internetgateway1"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/pkg/logger"
)

var (
	uPnPClientsIP       []*internetgateway1.WANIPConnection1
	uPnPClientsPPP      []*internetgateway1.WANPPPConnection1
	ipPortMappedClient  *internetgateway1.WANIPConnection1
	pppPortMappedClient *internetgateway1.WANPPPConnection1

	iface = flag.String("interface", "default",
		"network interface on which to perform connections. If not set, will use default interface.")

	// Log used for upnp
	Log *logger.ParanoidLogger
)

const attemptedPortAssignments = 10

// DiscoverDevices with UPnP on the network
func DiscoverDevices() error {
	var discoveryFinished sync.WaitGroup
	discoveryFinished.Add(2)

	go func() {
		defer discoveryFinished.Done()
		ipclients, _, err := internetgateway1.NewWANIPConnection1Clients()
		if err == nil {
			uPnPClientsIP = ipclients
		}
	}()

	go func() {
		defer discoveryFinished.Done()
		pppclients, _, err := internetgateway1.NewWANPPPConnection1Clients()
		if err == nil {
			uPnPClientsPPP = pppclients
		}
	}()

	discoveryFinished.Wait()
	if len(uPnPClientsIP) > 0 || len(uPnPClientsPPP) > 0 {
		return nil
	}
	return errors.New("No devices found")
}

func getUnoccupiedPortsIP(client *internetgateway1.WANIPConnection1) []int {
	m := make(map[int]bool)
	for i := 1; i < 65536; i++ {
		_, port, _, _, _, _, _, _, err := client.GetGenericPortMappingEntry(uint16(i))
		if err != nil {
			break
		}
		m[int(port)] = true
	}
	openPorts := make([]int, 0)
	for i := 1; i < 65536; i++ {
		if m[i] == false {
			openPorts = append(openPorts, i)
		}
	}
	return openPorts
}

func getUnoccupiedPortsppp(client *internetgateway1.WANPPPConnection1) []int {
	m := make(map[int]bool)
	for i := 1; i < 65536; i++ {
		_, port, _, _, _, _, _, _, err := client.GetGenericPortMappingEntry(uint16(i))
		if err != nil {
			break
		}
		m[int(port)] = true
	}
	openPorts := make([]int, 0)
	for i := 1; i < 65536; i++ {
		if m[i] == false {
			openPorts = append(openPorts, i)
		}
	}
	return openPorts
}

// AddPortMapping maps to a port
func AddPortMapping(internalIP string, internalPort int) (int, error) {
	for _, client := range uPnPClientsIP {
		openPorts := getUnoccupiedPortsIP(client)
		if len(openPorts) > 0 {
			for i := 0; i < attemptedPortAssignments; i++ {
				port := openPorts[rand.Intn(len(openPorts))]
				Log.Info("Picked port:", port)
				err := client.AddPortMapping("", uint16(port), "tcp", uint16(internalPort), internalIP, true, "", 0)
				if err == nil {
					ipPortMappedClient = client
					return port, nil
				}

				Log.Warn("Unable to map port", port, ". Error:", err)
			}
		}
	}
	for _, client := range uPnPClientsPPP {
		openPorts := getUnoccupiedPortsppp(client)
		if len(openPorts) > 0 {
			for i := 0; i < attemptedPortAssignments; i++ {
				port := openPorts[rand.Intn(len(openPorts))]
				Log.Info("Picked port:", port)
				err := client.AddPortMapping("", uint16(port), "tcp", uint16(internalPort), internalIP, true, "", 0)
				if err == nil {
					pppPortMappedClient = client
					return port, nil
				}

				Log.Warn("Unable to map port", port, ". Error:", err)
			}
		}
	}
	return 0, errors.New("Unable to map port")
}

// ClearPortMapping removes a UPnP port
func ClearPortMapping(externalPortString string) error {
	externalPort, err := strconv.Atoi(externalPortString)
	if err != nil {
		return err
	}
	if ipPortMappedClient != nil {
		return ipPortMappedClient.DeletePortMapping("", uint16(externalPort), "TCP")
	}
	if pppPortMappedClient != nil {
		return pppPortMappedClient.DeletePortMapping("", uint16(externalPort), "TCP")
	}
	return errors.New("No UPnP device available")
}

// GetInternalIP gets the internal Ip address
func GetInternalIP() (string, error) {
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		if (i.Flags & net.FlagLoopback) != 0 {
			continue
		}
		if *iface != "default" && i.Name != *iface {
			continue
		}
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("No matching interfaces found")
}

// GetExternalIP gets the external IP of the port mapped device.
func GetExternalIP() (string, error) {
	if ipPortMappedClient != nil {
		externalIP, err := ipPortMappedClient.GetExternalIPAddress()
		if err == nil {
			return externalIP, nil
		}
	}
	if pppPortMappedClient != nil {
		externalIP, err := pppPortMappedClient.GetExternalIPAddress()
		if err == nil {
			return externalIP, nil
		}
	}
	return "", errors.New("Unable to get get external IP address")
}

// GetIP gets an IP address that can be used
func GetIP() (string, error) {
	if globals.UPnPEnabled {
		return GetExternalIP()
	}

	return GetInternalIP()
}
