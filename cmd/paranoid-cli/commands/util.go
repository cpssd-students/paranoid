package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/user"
	"path"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/cpssd-students/paranoid/proto/paranoid/raft/v1"
)

type fileSystemAttributes struct {
	Encrypted    bool `json:"encrypted"`
	KeyGenerated bool `json:"keygenerated"`
	NetworkOff   bool `json:"networkoff"`
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func getRandomName() string {
	prefix := []string{
		"raging",
		"violent",
		"calm",
		"peaceful",
		"strange",
		"hungry",
	}
	postfix := []string{
		"dolphin",
		"snake",
		"elephant",
		"fox",
		"dog",
		"cat",
		"rabbit",
	}

	rand.Seed(time.Now().Unix())
	return prefix[rand.Int()%len(prefix)] + "_" + postfix[rand.Int()%len(postfix)]
}

// fileSystemExists checks if there is a folder in ~/.pfs with the given name
func fileSystemExists(fsname string) bool {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dirpath := path.Join(usr.HomeDir, ".pfs", "filesystems", fsname)
	_, err = ioutil.ReadDir(dirpath)
	return err == nil
}

// fileToProto converts a given file with a protobuf to a protobuf object
func fileToProto(file os.FileInfo, directory string) (entry *pb.LogEntry, err error) {
	filePath := path.Join(directory, file.Name())
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read logfile: %s", file.Name())
	}
	entry = &pb.LogEntry{
		Term:  0,
		Entry: nil,
	}
	err = proto.Unmarshal(fileData, entry)
	if err != nil {
		return nil, errors.New("Failed to Unmarshal file data")
	}
	return entry, nil
}
