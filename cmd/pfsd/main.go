package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"paranoid/pkg/libpfs"
	"paranoid/pkg/libpfs/encryption"
	"paranoid/pkg/logger"
	"paranoid/cmd/pfsd/dnetclient"
	"paranoid/cmd/pfsd/globals"
	"paranoid/cmd/pfsd/intercom"
	"paranoid/cmd/pfsd/keyman"
	"paranoid/cmd/pfsd/pfi"
	"paranoid/cmd/pfsd/pnetclient"
	"paranoid/cmd/pfsd/pnetserver"
	"paranoid/cmd/pfsd/upnp"
	pb "paranoid/pkg/proto/paranoidnetwork"
	rpb "paranoid/pkg/proto/raft"
	"paranoid/pkg/raft"
	"paranoid/pkg/raft/raftlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// GenerationJoinTimeout determines the timeout for joining a generation
	GenerationJoinTimeout time.Duration = time.Minute * 3
	// JoinSendKeysInterval determines the interval at which the keys should be
	// sent
	JoinSendKeysInterval time.Duration = time.Second
)

var (
	srv *grpc.Server
	log *logger.ParanoidLogger
)

// Flags
var (
	certFile = flag.String(
		"cert",
		"",
		"TLS certificate file - if empty connection will be unencrypted")
	keyFile = flag.String(
		"key",
		"",
		"TLS key file - if empty connection will be unencrypted")
	skipVerify = flag.Bool(
		"skip_verification",
		false,
		"skip verification of TLS certificate chain and hostname"+
			"- not recommended unless using self-signed certs")
	paranoidDirFlag = flag.String(
		"paranoid_dir",
		"",
		"Paranoid directory")
	mountDirFlag = flag.String(
		"mount_dir",
		"",
		"FUSE mount directory")
	discoveryAddrFlag = flag.String(
		"discovery_addr",
		"",
		"Address of discovery server")
	discoveryPoolFlag = flag.String(
		"discovery_pool",
		"",
		"pool to join")
	discoveryPasswordFlag = flag.String(
		"pool_password",
		"",
		"pool password")
)

type keySentResponse struct {
	err  error
	uuid string
}

func startKeyStateMachine() {
	_, err := os.Stat(path.Join(globals.ParanoidDir, "meta", keyman.KsmFileName))
	if err == nil {
		var err error
		keyman.StateMachine, err = keyman.NewKSMFromPFSDir(globals.ParanoidDir)
		if err != nil {
			log.Fatal("Unable to start key state machine:", err)
		}
	} else if os.IsNotExist(err) {
		keyman.StateMachine = keyman.NewKSM(globals.ParanoidDir)
	} else {
		log.Fatal("Error stating key state machine file")
	}
}

func sendKeyPiece(uuid string, generation int64, piece *keyman.KeyPiece, responseChan chan keySentResponse) {
	err := pnetclient.SendKeyPiece(uuid, generation, piece, true)
	responseChan <- keySentResponse{
		err:  err,
		uuid: uuid,
	}
}

func startRPCServer(lis *net.Listener, password string) {
	var opts []grpc.ServerOption
	if globals.TLSEnabled {
		log.Info("Starting ParanoidNetwork server with TLS.")
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatal("Failed to generate TLS credentials:", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	} else {
		log.Info("Starting ParanoidNetwork server without TLS.")
	}
	srv = grpc.NewServer(opts...)

	pb.RegisterParanoidNetworkServer(srv, &pnetserver.ParanoidServer{})
	nodeDetails := raft.Node{
		IP:         globals.ThisNode.IP,
		Port:       globals.ThisNode.Port,
		CommonName: globals.ThisNode.CommonName,
		NodeID:     globals.ThisNode.UUID,
	}

	startKeyStateMachine()

	if globals.Encrypted && globals.KeyGenerated {
		log.Info("Attempting to unlock")
		Unlock()
	}

	//First node to join a given cluster
	if len(globals.Nodes.GetAll()) == 0 {
		log.Info("Performing first node setup")
		globals.RaftNetworkServer = raft.NewNetworkServer(
			nodeDetails,
			globals.ParanoidDir,
			path.Join(globals.ParanoidDir, "meta", "raft"),
			&raft.StartConfiguration{
				Peers: []raft.Node{},
			},
			globals.TLSEnabled,
			globals.TLSSkipVerify,
			globals.Encrypted,
		)
		timeout := time.After(GenerationJoinTimeout)
	initalGenerationLoop:
		for {
			select {
			case <-timeout:
				log.Fatal("Unable to create initial generation")
			default:
				_, _, err := globals.RaftNetworkServer.RequestNewGeneration(globals.ThisNode.UUID)
				if err == nil {
					log.Info("Successfully created initial generation")
					break initalGenerationLoop
				}
				log.Error("Unable to create initial generation:", err)
			}
		}
		if globals.Encrypted {
			globals.KeyGenerated = true
			saveFileSystemAttributes(&globals.FileSystemAttributes{
				Encrypted:    globals.Encrypted,
				KeyGenerated: globals.KeyGenerated,
				NetworkOff:   globals.NetworkOff,
			})
		}
	} else {
		globals.RaftNetworkServer = raft.NewNetworkServer(
			nodeDetails,
			globals.ParanoidDir,
			path.Join(globals.ParanoidDir, "meta", "raft"),
			nil,
			globals.TLSEnabled,
			globals.TLSSkipVerify,
			globals.Encrypted,
		)
	}

	rpb.RegisterRaftNetworkServer(srv, globals.RaftNetworkServer)

	globals.Wait.Add(1)
	go func() {
		defer globals.Wait.Done()
		err := srv.Serve(*lis)
		log.Info("Paranoid network server stopped")
		if err != nil && globals.ShuttingDown == false {
			log.Fatal("Server stopped because of an error:", err)
		}
	}()

	if globals.Encrypted && !globals.KeyGenerated {
		timeout := time.After(GenerationJoinTimeout)
	generationCreateLoop:
		for {
			select {
			case <-timeout:
				log.Fatal("Unable to join cluster before timeout")
			default:
				generation, peers, err := pnetclient.NewGeneration(password)
				if err != nil {
					log.Error("Unable to start new generation:", err)
				}

				keyPiecesN := int64(len(peers) + 1)
				minKeysRequired := (keyPiecesN / 2) + 1
				log.Info("pieces : ", keyPiecesN)
				keyPieces, err := keyman.GeneratePieces(globals.EncryptionKey, keyPiecesN, minKeysRequired)
				if err != nil {
					log.Fatal("Unable to split keys:", err)
				}
				if len(keyPieces) != int(keyPiecesN) {
					log.Fatal("Unable to split keys: incorrect number of pieces returned. Got:", len(keyPieces),
						"Expected:", keyPiecesN)
				}

				err = globals.HeldKeyPieces.AddPiece(generation, globals.ThisNode.UUID, keyPieces[0])
				if err != nil {
					log.Fatal("Unable to store my key piece")
				}
				keyPieces = keyPieces[1:]

				log.Info("pieces : ", len(keyPieces))

				sendKeysTimer := time.NewTimer(0)
				sendKeysResponse := make(chan keySentResponse, len(peers))
				attemptJoin := make(chan bool, 100)
				keysReplicated := int64(1)
				var sendKeyPieceWait sync.WaitGroup

			sendKeysLoop:
				for {
					select {
					case <-timeout:
						log.Fatal("Unable to join cluster before timeout")
					case <-sendKeysTimer.C:
						for i := 0; i < len(peers); i++ {
							sendKeyPieceWait.Add(1)
							x := i
							go func() {
								defer sendKeyPieceWait.Done()
								sendKeyPiece(peers[x], generation, keyPieces[x], sendKeysResponse)
							}()
						}
						if keysReplicated >= minKeysRequired {
							attemptJoin <- true
						}
						sendKeysTimer.Reset(JoinSendKeysInterval)
					case keySendInfo := <-sendKeysResponse:
						log.Info("Received key piece response")
						if keySendInfo.err != nil {
							if keySendInfo.err == keyman.ErrGenerationDeprecated {
								log.Error("Attempting to replicate keys for deprecated generation")
								break sendKeysLoop
							} else {
								log.Error("Error sending key info:", keySendInfo.err)
							}
						} else {
							for i := 0; i < len(peers); i++ {
								if peers[i] == keySendInfo.uuid {
									peers = append(peers[:i], peers[i+1:]...)
									keyPieces = append(keyPieces[:i], keyPieces[i+1:]...)
									keysReplicated++
									if keysReplicated >= minKeysRequired {
										attemptJoin <- true
									}
								}
							}
						}
					case <-attemptJoin:
						log.Info("Attempting to join raft cluster")
						err := pnetclient.JoinCluster(password)
						if err != nil {
							log.Error("Unable to join a raft cluster:", err)
						} else {
							log.Info("Successfully joined raft cluster")
							globals.Wait.Add(1)
							go func() {
								defer globals.Wait.Done()
								done := make(chan bool, 1)
								go func() {
									sendKeyPieceWait.Wait()
									done <- true
								}()
								for {
									select {
									case <-sendKeysResponse:
									case <-done:
										return
									}
								}
							}()
							break generationCreateLoop
						}
					}
				}
			}
		}

		globals.KeyGenerated = true
		saveFileSystemAttributes(&globals.FileSystemAttributes{
			Encrypted:    globals.Encrypted,
			KeyGenerated: globals.KeyGenerated,
			NetworkOff:   globals.NetworkOff,
		})
	} else if globals.RaftNetworkServer.State.Configuration.HasConfiguration() == false {
		log.Info("Attempting to join raft cluster")
		err := dnetclient.JoinCluster(password)
		if err != nil {
			log.Fatal("Unable to join a raft cluster")
		}
	}

	globals.Wait.Add(1)
	go pnetclient.KSMObserver(keyman.StateMachine)
}

func setupLogging() {
	logDir := path.Join(globals.ParanoidDir, "meta", "logs")

	log = logger.New("main", "pfsd", logDir)
	dnetclient.Log = logger.New("dnetclient", "pfsd", logDir)
	pnetclient.Log = logger.New("pnetclient", "pfsd", logDir)
	pnetserver.Log = logger.New("pnetserver", "pfsd", logDir)
	upnp.Log = logger.New("upnp", "pfsd", logDir)
	keyman.Log = logger.New("keyman", "pfsd", logDir)
	raft.Log = logger.New("raft", "pfsd", logDir)
	raftlog.Log = logger.New("raftlog", "pfsd", logDir)
	libpfs.Log = logger.New("libpfs", "pfsd", logDir)
	intercom.Log = logger.New("intercom", "pfsd", logDir)
	globals.Log = logger.New("globals", "pfsd", logDir)

	log.SetOutput(logger.STDERR | logger.LOGFILE)
	dnetclient.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	pnetclient.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	pnetserver.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	upnp.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	keyman.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	raft.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	raftlog.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	libpfs.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	intercom.Log.SetOutput(logger.STDERR | logger.LOGFILE)
	globals.Log.SetOutput(logger.STDERR | logger.LOGFILE)
}

func getFileSystemAttributes() {
	attributesJSON, err := ioutil.ReadFile(path.Join(globals.ParanoidDir, "meta", "attributes"))
	if err != nil {
		log.Fatal("unable to read file system attributes:", err)
	}

	attributes := &globals.FileSystemAttributes{}
	err = json.Unmarshal(attributesJSON, attributes)
	if err != nil {
		log.Fatal("unable to read file system attributes:", err)
	}

	globals.Encrypted = attributes.Encrypted
	globals.NetworkOff = attributes.NetworkOff
	encryption.Encrypted = attributes.Encrypted

	if attributes.Encrypted {
		if !attributes.KeyGenerated {
			//If a key has not yet been generated for this file system, one must be generated
			globals.EncryptionKey, err = keyman.GenerateKey(32)
			if err != nil {
				log.Fatal("unable to generate encryption key:", err)
			}

			cipherB, err := encryption.GenerateAESCipherBlock(globals.EncryptionKey.GetBytes())
			if err != nil {
				log.Fatal("unable to generate cipher block:", err)
			}
			encryption.SetCipher(cipherB)

			if attributes.NetworkOff {
				//If networking is turned off, save the key to a file
				attributes.KeyGenerated = true
				attributes.EncryptionKey = *globals.EncryptionKey
			}
		} else if attributes.NetworkOff {
			//If networking is off, load the key from the file
			globals.EncryptionKey = &attributes.EncryptionKey
			cipherB, err := encryption.GenerateAESCipherBlock(globals.EncryptionKey.GetBytes())
			if err != nil {
				log.Fatal("unable to generate cipher block:", err)
			}
			encryption.SetCipher(cipherB)
		}
	}

	globals.KeyGenerated = attributes.KeyGenerated
	if globals.KeyGenerated {
		LoadPieces()
	}

	saveFileSystemAttributes(attributes)
}

func saveFileSystemAttributes(attributes *globals.FileSystemAttributes) {
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		log.Fatal("unable to save new file system attributes to file:", err)
	}

	newAttributesFile := path.Join(globals.ParanoidDir, "meta", "attributes-new")
	err = ioutil.WriteFile(newAttributesFile, attributesJSON, 0600)
	if err != nil {
		log.Fatal("unable to save new file system attributes to file:", err)
	}

	err = os.Rename(newAttributesFile, path.Join(globals.ParanoidDir, "meta", "attributes"))
	if err != nil {
		log.Fatal("unable to save new file system attributes to file:", err)
	}
}

func main() {
	flag.Parse()

	var err error

	if len(*paranoidDirFlag) == 0 {
		fmt.Println("FATAL: paranoid directory must be provided")
		os.Exit(1)
	}
	globals.ParanoidDir, err = filepath.Abs(*paranoidDirFlag)
	if err != nil {
		fmt.Println("FATAL: Could not get absolute paranoid dir:", err)
		os.Exit(1)
	}

	if len(*mountDirFlag) == 0 {
		fmt.Println("FATAL: mount point must be provided")
		os.Exit(1)
	}
	globals.MountPoint, err = filepath.Abs(*mountDirFlag)
	if err != nil {
		fmt.Println("FATAL: Could not get absolute mount point:", err)
		os.Exit(1)
	}

	setupLogging()
	getFileSystemAttributes()

	globals.TLSSkipVerify = *skipVerify
	if *certFile != "" && *keyFile != "" {
		globals.TLSEnabled = true
		if !globals.TLSSkipVerify {
			cn, err := getCommonNameFromCert(*certFile)
			if err != nil {
				log.Fatal("Could not get CN from provided TLS cert:", err)
			}
			globals.ThisNode.CommonName = cn
		}
	} else {
		globals.TLSEnabled = false
	}

	if !globals.NetworkOff {
		uuid, err := ioutil.ReadFile(path.Join(globals.ParanoidDir, "meta", "uuid"))
		if err != nil {
			log.Fatal("Could not get node UUID:", err)
		}
		globals.ThisNode.UUID = string(uuid)

		ip, err := upnp.GetIP()
		if err != nil {
			log.Fatal("Could not get IP:", err)
		}

		//Asking for port 0 requests a random free port from the OS.
		lis, err := net.Listen("tcp", ip+":0")
		if err != nil {
			log.Fatalf("Failed to start listening : %v.\n", err)
		}
		splits := strings.Split(lis.Addr().String(), ":")
		port := splits[len(splits)-1]
		portInt, err := strconv.Atoi(port)
		if err != nil {
			log.Fatal("Could not parse port", splits[len(splits)-1], " Error :", err)
		}
		globals.ThisNode.Port = port

		//Try and set up uPnP. Otherwise use internal IP.
		globals.UPnPEnabled = false
		err = upnp.DiscoverDevices()
		if err == nil {
			log.Info("UPnP devices available")
			externalPort, err := upnp.AddPortMapping(ip, portInt)
			if err == nil {
				log.Info("UPnP port mapping enabled")
				port = strconv.Itoa(externalPort)
				globals.ThisNode.Port = port
				globals.UPnPEnabled = true
			}
		}

		globals.ThisNode.IP, err = upnp.GetIP()
		if err != nil {
			log.Fatal("Can't get IP. Error : ", err)
		}
		log.Info("Peer address:", globals.ThisNode.IP+":"+globals.ThisNode.Port)

		if _, err := os.Stat(globals.ParanoidDir); os.IsNotExist(err) {
			log.Fatal("Path", globals.ParanoidDir, "does not exist.")
		}
		if _, err := os.Stat(path.Join(globals.ParanoidDir, "meta")); os.IsNotExist(err) {
			log.Fatal("Path", globals.ParanoidDir, "is not valid PFS root.")
		}

		if len(*discoveryAddrFlag) == 0 {
			log.Fatal("discovery server address must be specified")
		}
		dnetclient.SetDiscovery(*discoveryAddrFlag)
		dnetclient.JoinDiscovery(*discoveryPoolFlag, *discoveryPasswordFlag)
		if err = globals.SetPoolPasswordHash(*discoveryPasswordFlag); err != nil {
			log.Fatal("Error setting up password hash:", err)
		}
		startRPCServer(&lis, *discoveryPasswordFlag)
	}
	createPid("pfsd")
	pfi.StartPfi(false)

	intercom.RunServer(path.Join(globals.ParanoidDir, "meta"))

	HandleSignals()
}

func createPid(processName string) {
	processID := os.Getpid()
	pid := []byte(strconv.Itoa(processID))
	err := ioutil.WriteFile(path.Join(globals.ParanoidDir, "meta", processName+".pid"), pid, 0600)
	if err != nil {
		log.Fatal("Failed to create PID file", err)
	}
}
