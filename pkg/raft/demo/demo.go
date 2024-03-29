package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/cpssd-students/paranoid/pkg/raft"
	"github.com/cpssd-students/paranoid/pkg/raft/rafttestutil"
	pb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

// Constants used in the demo
const (
	DemoDuration          time.Duration = 50 * time.Second
	PrintTime                           = 5 * time.Second
	MinRandomNumberGen                  = 3000 * time.Millisecond
	MaxRandomNumberGen                  = 10000 * time.Millisecond
	MinRandomDropInterval               = 5000 * time.Millisecond
	MaxRandomDropInterval               = 10000 * time.Millisecond
)

// TODO: Either use the MaxRandomNumberGen or remove it
var _ = MaxRandomNumberGen

var (
	waitGroup sync.WaitGroup
	demo      = flag.Int("demo", 0, "Which demo to run (1-3). 0 for all demos")
)

func getRandomInterval() time.Duration {
	interval := int64(MaxRandomDropInterval) - int64(MinRandomDropInterval)
	randx := rand.Int63n(interval)
	return MinRandomNumberGen + time.Duration(randx)
}

func getRandomDropInterval() time.Duration {
	interval := int64(MaxRandomDropInterval) - int64(MinRandomDropInterval)
	randx := rand.Int63n(interval)
	return MinRandomDropInterval + time.Duration(randx)
}

func getRandomDrop() time.Duration {
	randomChance := rand.Intn(100)
	if randomChance < 20 { //5 to 8 second drop
		interval := int64(8*time.Second) - int64(5*time.Second)
		randx := rand.Int63n(interval)
		return time.Second*5 + time.Duration(randx)
	}
	// 2 to 5 second drop
	interval := int64(5*time.Second) - int64(2*time.Second)
	randx := rand.Int63n(interval)
	return time.Second*2 + time.Duration(randx)
}

func manageNode(raftServer *raft.NetworkServer) {
	defer waitGroup.Done()
	testTimer := time.NewTimer(DemoDuration)
	randomNumTimer := time.NewTimer(getRandomInterval())
	for {
		select {
		case <-testTimer.C:
			return
		case <-randomNumTimer.C:
			if raftServer.QuitChannelClosed {
				return
			}
			randomNumber := rand.Intn(1000)
			log.Printf("%s requesting that %d be added to the log\n",
				raftServer.State.NodeID, randomNumber)
			_, err := raftServer.RequestAddLogEntry(&pb.Entry{
				Type: pb.EntryType_ENTRY_TYPE_DEMO,
				Uuid: rafttestutil.GenerateNewUUID(),
				Demo: &pb.DemoCommand{Number: uint64(randomNumber)},
			})
			if err == nil {
				log.Printf("%s successfully added %d to the log\n",
					raftServer.State.NodeID, randomNumber)
			} else {
				log.Printf("%s could not add %d to the log: %v\n",
					raftServer.State.NodeID, randomNumber, err)
			}
			randomNumTimer.Reset(getRandomInterval())
		}
	}
}

func printLogs(cluster []*raft.NetworkServer) {
	defer waitGroup.Done()
	testTimer := time.NewTimer(DemoDuration)
	printTimer := time.NewTimer(PrintTime)
	for {
		select {
		case <-testTimer.C:
			return
		case <-printTimer.C:
			printTimer.Reset(PrintTime)
			log.Println("Printing node logs:")
			for i := 0; i < len(cluster); i++ {
				logsString := ""
				for j := uint64(1); j <= cluster[i].State.Log.GetMostRecentIndex(); j++ {
					logEntry, err := cluster[i].State.Log.GetLogEntry(j)
					if err != nil {
						log.Fatalln("Error reading log entry:", err)
					}
					logsString = fmt.Sprintf("%s %d", logsString, logEntry.Entry.GetDemo().Number)
				}
				log.Println(cluster[i].State.NodeID, "Logs:", logsString)
			}
		}
	}
}

func performDemoOne(node1RaftServer, node2RaftServer, node3RaftServer *raft.NetworkServer) {
	log.Println("Running basic raft demo")
	waitGroup.Add(4)
	go manageNode(node1RaftServer)
	go manageNode(node2RaftServer)
	go manageNode(node3RaftServer)
	go printLogs([]*raft.NetworkServer{node1RaftServer, node2RaftServer, node3RaftServer})

	waitGroup.Wait()
}

func performDemoTwo(
	node1srv *grpc.Server,
	node1RaftServer, node2RaftServer, node3RaftServer *raft.NetworkServer,
) {
	log.Println("Running raft node crash demo")
	waitGroup.Add(4)
	go manageNode(node1RaftServer)
	go manageNode(node2RaftServer)
	go manageNode(node3RaftServer)
	go printLogs([]*raft.NetworkServer{node1RaftServer, node2RaftServer, node3RaftServer})

	//Node1 will crash after 20 seconds
	time.Sleep(20 * time.Second)
	log.Println("Crashing node1")
	node1srv.Stop()
	rafttestutil.StopRaftServer(node1RaftServer)

	waitGroup.Wait()
}

func bringBackUp(
	currentlyDown []bool,
	nodePorts []string,
	nodeServers []*grpc.Server,
	nodeListners []*net.Listener,
	raftServers []*raft.NetworkServer,
	nodeNum int,
) {
	rafttestutil.CloseListener(nodeListners[nodeNum])
	time.Sleep(getRandomDrop())
	rafttestutil.CloseListener(nodeListners[nodeNum])
	lis, err := net.Listen("tcp", ":"+nodePorts[nodeNum])
	if err != nil {
		//log.Printf("Failed to start listening. Retrying later : %v.\n", err)
		bringBackUp(currentlyDown, nodePorts, nodeServers, nodeListners, raftServers, nodeNum)
		return
	}
	log.Println(raftServers[nodeNum].State.NodeID, "coming back up")
	nodeListners[nodeNum] = &lis
	go func() { _ = nodeServers[nodeNum].Serve(lis) }()
	currentlyDown[nodeNum] = false
}

func performDemoThree(
	nodePorts []string,
	nodeServers []*grpc.Server,
	nodeListners []*net.Listener,
	raftServers []*raft.NetworkServer,
) {
	log.Println("Running raft message drop demo")
	waitGroup.Add(4)
	go manageNode(raftServers[0])
	go manageNode(raftServers[1])
	go manageNode(raftServers[2])
	go printLogs(raftServers)

	nodeDowns := make([]bool, 3)
	testTimer := time.NewTimer(DemoDuration)
	nodeDownTimer := time.NewTimer(getRandomDropInterval())
	for {
		select {
		case <-testTimer.C:
			goto end
		case <-nodeDownTimer.C:
			nodeDown := rand.Intn(3)
			if !nodeDowns[nodeDown] {
				nodeDowns[nodeDown] = true
				log.Println(raftServers[nodeDown].State.NodeID, "dropping messages")
				rafttestutil.CloseListener(nodeListners[nodeDown])
				go bringBackUp(
					nodeDowns, nodePorts, nodeServers, nodeListners, raftServers, nodeDown)
			}
			nodeDownTimer.Reset(getRandomDropInterval())
		}
	}

end:
	waitGroup.Wait()
}

func setupDemo(demoNum int) {
	node1Lis, node1Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node1Lis)
	node1 := rafttestutil.SetUpNode("node1", "localhost", node1Port, "_")
	node2Lis, node2Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node2Lis)
	node2 := rafttestutil.SetUpNode("node2", "localhost", node2Port, "_")
	node3Lis, node3Port := rafttestutil.StartListener()
	defer rafttestutil.CloseListener(node3Lis)
	node3 := rafttestutil.SetUpNode("node3", "localhost", node3Port, "_")

	node1RaftDirectory := rafttestutil.CreateRaftDirectory(
		path.Join(os.TempDir(), "rafttest1", "node1"))
	var node1RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node1RaftDirectory, node1RaftServer)
	node1RaftServer, node1srv := raft.StartRaft(
		node1Lis,
		node1,
		"",
		node1RaftDirectory,
		&raft.StartConfiguration{Peers: []raft.Node{node2, node3}},
	)
	defer node1srv.Stop()
	defer rafttestutil.StopRaftServer(node1RaftServer)

	node2RaftDirectory := rafttestutil.CreateRaftDirectory(
		path.Join(os.TempDir(), "rafttest1", "node2"))
	var node2RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node2RaftDirectory, node2RaftServer)
	node2RaftServer, node2srv := raft.StartRaft(
		node2Lis,
		node2,
		"",
		node2RaftDirectory,
		&raft.StartConfiguration{Peers: []raft.Node{node1, node3}},
	)
	defer node2srv.Stop()
	defer rafttestutil.StopRaftServer(node2RaftServer)

	node3RaftDirectory := rafttestutil.CreateRaftDirectory(
		path.Join(os.TempDir(), "rafttest1", "node3"))
	var node3RaftServer *raft.NetworkServer
	defer rafttestutil.RemoveRaftDirectory(node3RaftDirectory, node3RaftServer)
	node3RaftServer, node3srv := raft.StartRaft(
		node3Lis,
		node3,
		"",
		node3RaftDirectory,
		&raft.StartConfiguration{Peers: []raft.Node{node1, node2}},
	)
	defer node3srv.Stop()
	defer rafttestutil.StopRaftServer(node3RaftServer)

	if demoNum == 1 {
		performDemoOne(node1RaftServer, node2RaftServer, node3RaftServer)
	} else if demoNum == 2 {
		performDemoTwo(node1srv, node1RaftServer, node2RaftServer, node3RaftServer)
	} else if demoNum == 3 {
		performDemoThree([]string{node1Port, node2Port, node2Port},
			[]*grpc.Server{node1srv, node2srv, node3srv},
			[]*net.Listener{node1Lis, node2Lis, node3Lis},
			[]*raft.NetworkServer{node1RaftServer, node2RaftServer, node3RaftServer})
	}
}

func createDemoDirectory() {
	removeDemoDirectory()
	err := os.Mkdir(path.Join(os.TempDir(), "raftdemo"), 0777)
	if err != nil {
		log.Fatalln("Error creating demo directory:", err)
	}
}

func removeDemoDirectory() {
	os.RemoveAll(path.Join(os.TempDir(), "raftdemo"))
}

func createFileWriter(logPath string, component string) (io.Writer, error) {
	return os.OpenFile(path.Join(logPath, component+".log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	createDemoDirectory()
	defer removeDemoDirectory()

	writer, err := createFileWriter(path.Join(os.TempDir(), "raftdemo"), "grpclog")
	if err != nil {
		log.Println("Could not redirect grpc output")
	} else {
		grpclog.SetLoggerV2(grpclog.NewLoggerV2(writer, writer, writer))
	}

	flag.Parse()

	demo := *demo
	if demo == 0 {
		setupDemo(1)
		setupDemo(2)
		setupDemo(3)
	} else {
		if demo > 3 || demo < 1 {
			log.Fatal("Bad demo number specified")
		}
		setupDemo(demo)
	}
}
