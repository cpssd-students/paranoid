package libpfs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

//makeDir creates a new directory with permissions 0777 with the name newDir in parentDir.
func makeDir(parentDir, newDir string) (string, error) {
	newDirPath := path.Join(parentDir, newDir)
	err := os.Mkdir(newDirPath, 0700)
	if err != nil {
		return "", err
	}
	return newDirPath, nil
}

//checkEmpty checks if a given directory has any children.
func checkEmpty(directory string) error {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("error reading directory %s", err)
	}
	if len(files) > 0 {
		return errors.New("init : directory must be empty")
	}
	return nil
}

//InitCommand creates the pvd directory sturucture
//It also gets a UUID and stores it in the meta directory.
func InitCommand(directory string) (returnCode returncodes.Code, returnError error) {
	log.Printf("creating new pfs in %s", directory)

	err := checkEmpty(directory)
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	_, err = makeDir(directory, "names")
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	_, err = makeDir(directory, "inodes")
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	metaDir, err := makeDir(directory, "meta")
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	_, err = makeDir(metaDir, "logs")
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	_, err = makeDir(metaDir, "raft")
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	_, err = makeDir(directory, "contents")
	if err != nil {
		return returncodes.EUNEXPECTED, err
	}

	uuid, err := ioutil.ReadFile("/proc/sys/kernel/random/uuid")
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error reading uuid: %s", err)
	}

	uuidString := strings.TrimSpace(string(uuid))
	log.Printf("%s init UUID: %s", directory, uuidString)

	err = ioutil.WriteFile(path.Join(metaDir, "uuid"), []byte(uuidString), 0600)
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error writing uuid file: %s", err)
	}

	_, err = os.Create(path.Join(metaDir, "lock"))
	if err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error creating lock file: %s", err)
	}
	return returncodes.OK, nil
}
