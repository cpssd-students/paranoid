package globals

import (
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pp2p/paranoid/logger"
	"github.com/pp2p/paranoid/pfsd/keyman"
	"github.com/pp2p/paranoid/raft"
	"golang.org/x/crypto/bcrypt"
)

const (
	// PasswordSaltLength is the length of the password salt
	PasswordSaltLength int = 64
)

// Log is used to log messages from pfsd
var Log *logger.ParanoidLogger

// Node struct
type Node struct {
	IP         string
	Port       string
	CommonName string
	UUID       string
}

func (n Node) String() string {
	return fmt.Sprintf("%s:%s", n.IP, n.Port)
}

// FileSystemAttributes stores the information written onto the disk
type FileSystemAttributes struct {
	Encrypted     bool       `json:"encrypted"`
	KeyGenerated  bool       `json:"keygenerated"`
	NetworkOff    bool       `json:"networkoff"`
	EncryptionKey keyman.Key `json:"encryptionkey,omitempty"` //The encryption key is only saved to file in this manner if networking is turned off
}

// RaftNetworkServer is an instance of the network server
var RaftNetworkServer *raft.RaftNetworkServer

// ParanoidDir is the location of the stored encrypted data
var ParanoidDir string

// MountPoint of the FUSE filesystem
var MountPoint string

// BootTime at which PFSD started. Used for calculating uptime.
var BootTime time.Time

// ThisNode stores information about the current node
var ThisNode Node

// UPnPEnabled is kinda self explanatory
var UPnPEnabled bool

// ResetInterval containing how often the Renew function has to be called
var ResetInterval time.Duration

// DiscoveryAddr contains the connection sting to the discovery server
var DiscoveryAddr string

// Nodes instance which controls all the information about other pfsd instances
var Nodes = nodes{m: make(map[string]Node)}

// NetworkOff if the network should not be used
// TODO: Rename it to NetworkActive
var NetworkOff bool

// TLSEnabled if network encryption is used in all connections to and from PFSD
var TLSEnabled bool

// TLSSkipVerify will cause PFSD to not verify the TLS credentials of anything
// it connects to
var TLSSkipVerify bool

// PoolPasswordHash used to connect to the pool
var PoolPasswordHash []byte

// PoolPasswordSalt used to connect to the pool
var PoolPasswordSalt []byte

// SetPoolPasswordHash generates and sets a new password hash from the password
func SetPoolPasswordHash(password string) error {
	PoolPasswordHash = make([]byte, 0)
	PoolPasswordSalt = make([]byte, PasswordSaltLength)
	n, err := io.ReadFull(rand.Reader, PoolPasswordSalt)
	if err != nil {
		return err
	}
	if n != PasswordSaltLength {
		return errors.New("unable to read correct number of bytes from random number generator")
	}

	if password != "" {
		PoolPasswordHash, err = bcrypt.GenerateFromPassword(append(PoolPasswordSalt, []byte(password)...), bcrypt.DefaultCost)
		return err
	}
	return nil
}

// Wait for all goroutines in PFSD
var Wait sync.WaitGroup

// Quit channel used for killing the application
// TODO: Use struct{} instead
var Quit = make(chan bool)

// ShuttingDown is set when the PFSD is in shutdown phase
var ShuttingDown bool

// --------------------------------------------
// ---- nodes ---- //
// --------------------------------------------

type nodes struct {
	m    map[string]Node
	lock sync.Mutex
}

func (ns *nodes) Add(n Node) {
	ns.lock.Lock()
	defer ns.lock.Unlock()
	ns.m[n.UUID] = n
}

func (ns *nodes) GetNode(uuid string) (Node, error) {
	ns.lock.Lock()
	defer ns.lock.Unlock()
	node, ok := ns.m[uuid]
	if !ok {
		return node, errors.New("unrecognised uuid")
	}
	return node, nil
}

func (ns *nodes) Remove(n Node) {
	ns.lock.Lock()
	defer ns.lock.Unlock()
	delete(ns.m, n.UUID)
}

func (ns *nodes) GetAll() []Node {
	ns.lock.Lock()
	defer ns.lock.Unlock()

	var res []Node
	for _, node := range ns.m {
		res = append(res, node)
	}
	return res
}

//	--------------------
//	---- Encryption ----
//	--------------------

// Encrypted determines whether the whole filesystem is encrypted
var Encrypted bool

// KeyGenerated is set to true when the key is generated
var KeyGenerated bool

// EncryptionKey stores the actual key used for encryption
var EncryptionKey *keyman.Key

var keyPieceStoreLock sync.Mutex

// KeyPieceMap of the key pieces
type KeyPieceMap map[string]*keyman.KeyPiece

// KeyPieceStore maps key generations to individual key pieces
type KeyPieceStore map[int64]KeyPieceMap

// GetPiece gets the key piece based on the generation and the UUID of the node.
// It returns nil if the key is not found.
func (ks KeyPieceStore) GetPiece(generation int64, nodeUUID string) *keyman.KeyPiece {
	keyPieceStoreLock.Lock()
	defer keyPieceStoreLock.Unlock()

	keymap, ok := ks[generation]
	if !ok {
		return nil
	}

	piece, ok := keymap[nodeUUID]
	if !ok {
		return nil
	}
	return piece
}

// AddPiece adds a specific piece of the key, associated with the node to a
// generation
func (ks KeyPieceStore) AddPiece(generation int64, nodeUUID string, piece *keyman.KeyPiece) error {
	keyPieceStoreLock.Lock()
	defer keyPieceStoreLock.Unlock()

	_, ok := ks[generation]
	if !ok {
		ks[generation] = make(KeyPieceMap)
	}

	ks[generation][nodeUUID] = piece
	return ks.SaveToDisk()
}

// DeletePiece for a node from a given generation
func (ks KeyPieceStore) DeletePiece(generation int64, nodeUUID string) error {
	keyPieceStoreLock.Lock()
	defer keyPieceStoreLock.Unlock()

	_, ok := ks[generation]
	if !ok {
		return nil
	}

	delete(ks[generation], nodeUUID)
	return ks.SaveToDisk()
}

// DeleteGeneration removes the whole generation
func (ks KeyPieceStore) DeleteGeneration(generation int64) error {
	keyPieceStoreLock.Lock()
	defer keyPieceStoreLock.Unlock()
	delete(ks, generation)
	return ks.SaveToDisk()
}

// SaveToDisk saves all the keypieces in the meta directory
func (ks KeyPieceStore) SaveToDisk() error {
	piecePath := path.Join(ParanoidDir, "meta", "pieces-new")
	file, err := os.Create(piecePath)
	if err != nil {
		Log.Errorf("Unable to open %s for storing pieces: %s", piecePath, file)
		return fmt.Errorf("Unable to open %s for storing pieces: %s", piecePath, file)
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	err = enc.Encode(ks)
	if err != nil {
		Log.Error("Failed encoding KeyPieceStore to GOB:", err)
		return fmt.Errorf("failed encoding KeyPieceStore to GOB: %s", err)
	}
	err = os.Rename(piecePath, path.Join(ParanoidDir, "meta", "pieces"))
	if err != nil {
		Log.Error("Failed to save KeyPieceStore to file:", err)
		return fmt.Errorf("Failed to save KeyPieceStore to file: %s", err)
	}
	return nil
}

// HeldKeyPieces is an instance of KeyPieceStore
var HeldKeyPieces = make(KeyPieceStore)
