package libpfs

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"

	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

//MountCommand is used to notify a pfs paranoidDirectory it has been mounted.
func MountCommand(
	paranoidDirectory, dAddr, mountPoint string,
) (returnCode returncodes.Code, returnError error) {
	log.Printf("mount %s in %s", paranoidDirectory, mountPoint)

	if err := ioutil.WriteFile(
		path.Join(paranoidDirectory, "meta", "discovery_address"),
		[]byte(dAddr),
		0600,
	); err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error saving discovery_address file: %s", err)
	}

	mountPoint, err := filepath.Abs(mountPoint)
	if err != nil {
		return returncodes.EUNEXPECTED,
			fmt.Errorf("error getting absolute path of mountpoint: %w", err)
	}

	if err := ioutil.WriteFile(
		path.Join(paranoidDirectory, "meta", "mountpoint"),
		[]byte(mountPoint),
		0600,
	); err != nil {
		return returncodes.EUNEXPECTED, fmt.Errorf("error writing mountpoint: %s", err)
	}

	return returncodes.OK, nil
}
