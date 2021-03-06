package dnetservertest

import (
	syslog "log"
	"os"
	"path"
	"strconv"
	"testing"

	. "paranoid/cmd/discovery-server/dnetserver"
	"paranoid/pkg/logger"
	pb "paranoid/pkg/proto/discoverynetwork"
)

func TestMain(m *testing.M) {
	Log = logger.New("discoveryTest", "discoveryTest", "/dev/null")
	Log.SetLogLevel(logger.ERROR)
	Pools = make(map[string]*Pool)
	StateDirectoryPath = path.Join(os.TempDir(), "server_state_bench")
	err := os.RemoveAll(StateDirectoryPath)
	if err != nil {
		Log.Fatal("Test setup failed:", err)
	}
	err = os.Mkdir(StateDirectoryPath, 0700)
	if err != nil {
		Log.Fatal("Test setup failed:", err)
	}
	os.Exit(m.Run())
}

func BenchmarkJoin(b *testing.B) {
	discovery := DiscoveryServer{}
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		joinRequest := pb.JoinRequest{
			Node: &pb.Node{CommonName: "TestNode" + str, Ip: "1.1.1." + str, Port: "1001", Uuid: "blahblah" + str},
			Pool: "TestPool" + strconv.Itoa(n/5),
		}
		_, err := discovery.Join(nil, &joinRequest)
		if err != nil {
			syslog.Fatalln("Error joining network : ", err)
		}
	}
}

func BenchmarkDisco(b *testing.B) {
	discovery := DiscoveryServer{}
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		joinRequest := pb.JoinRequest{
			Node: &pb.Node{CommonName: "TestNode" + str, Ip: "1.1.1.1" + str, Port: "1001", Uuid: "blahblah"},
			Pool: "TestPool" + strconv.Itoa(n/5),
		}
		discovery.Join(nil, &joinRequest)
		disconnect := pb.DisconnectRequest{
			Node: &pb.Node{CommonName: "TestNode" + str, Ip: "1.1.1.1" + str, Port: "1001", Uuid: "blahblah"},
			Pool: "TestPool" + strconv.Itoa(n/5),
		}
		_, err := discovery.Disconnect(nil, &disconnect)
		if err != nil {
			syslog.Fatalln("Error disconnecting node 2:", err)
		}
	}
}
