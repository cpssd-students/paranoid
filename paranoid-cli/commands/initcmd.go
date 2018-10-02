package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/urfave/cli"

	"paranoid/libpfs"
	"paranoid/libpfs/returncodes"
	log "paranoid/logger"
	"paranoid/paranoid-cli/tls"
)

func cleanupPFS(pfsDir string) {
	err := os.RemoveAll(pfsDir)
	if err != nil {
		log.Warn("Could not successfully clean up PFS directory.")
	}
}

// Init subcommand
func Init(c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		cli.ShowCommandHelp(c, "init")
		os.Exit(0)
	}

	doInit(
		args[0],
		c.String("pool"),
		c.String("cert"),
		c.String("key"),
		c.Bool("unsecure"),
		c.Bool("unencrypted"),
		c.Bool("networkoff"),
	)
}

//doInit inits a new paranoid file system
func doInit(pfsname, pool, cert, key string, unsecure, unencrypted, networkoff bool) {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Error Getting Current User")
		log.Fatalf("cannot get curent user: %v", err)
	}
	homeDir := usr.HomeDir

	if _, err = os.Stat(path.Join(homeDir, ".pfs", "filesystems")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(homeDir, ".pfs", "filesystems"), 0700)
		if err != nil {
			fmt.Println("FATAL: Error making pfs directory")
			log.Fatalf("unable to create pfs directory: %v", err)
		}
	}

	directory, err := filepath.Abs(path.Join(homeDir, ".pfs", "filesystems", pfsname))
	if err != nil {
		fmt.Println("FATAL: Given pfs-name is in incorrect format.")
		log.Fatalf("given pfs-name is in incorrect format. Error: %v", err)
	}
	if path.Base(directory) != pfsname {
		fmt.Println("FATAL: Given pfs-name is in incorrect format.")
		log.Fatal("given pfs-name is in incorrect format.")
	}

	if _, err = os.Stat(directory); !os.IsNotExist(err) {
		fmt.Println("FATAL: A paranoid file system with that name already exists")
		log.Fatal("paranoid file system with that name already exists")
	}
	err = os.Mkdir(directory, 0700)
	if err != nil {
		fmt.Println("FATAL: Unable to make pfs Directory")
		log.Fatalf("Error making pfs directory: %v", err)
	}

	returncode, err := libpfs.InitCommand(directory)
	if returncode != returncodes.OK {
		cleanupPFS(directory)
		fmt.Println("FATAL: Error running pfs init")
		log.Fatalf("Error running pfs init: %v", err)
	}

	// Either create a new pool name or use the one for a flag and save to meta/pool
	if len(pool) == 0 {
		pool = getRandomName()
	}
	err = ioutil.WriteFile(path.Join(directory, "meta", "pool"), []byte(pool), 0600)
	if err != nil {
		fmt.Println("FATAL: Cannot save pool information")
		log.Fatalf("cannot save pool information: %v", err)
	}
	fmt.Println("Using pool name", pool)
	log.Infof("Using pool name %s", pool)

	fileAttributes := &fileSystemAttributes{
		Encrypted:    !unencrypted,
		KeyGenerated: false,
		NetworkOff:   networkoff,
	}

	attributesJSON, err := json.Marshal(fileAttributes)
	if err != nil {
		fmt.Println("FATAL: Cannot save file system information")
		log.Fatalf("cannot save file system information: %v", err)
	}

	err = ioutil.WriteFile(path.Join(directory, "meta", "attributes"), attributesJSON, 0600)
	if err != nil {
		fmt.Println("FATAL: Cannot save file system information")
		log.Fatalf("cannot save file system information: %v", err)
	}

	if networkoff {
		return
	}

	if unsecure {
		fmt.Println("--unsecure specified. PFSD will not use TLS for its communication.")
		return
	}

	if (cert != "") && (key != "") {
		fmt.Println("Using existing certificate.")
		err = os.Link(cert, path.Join(directory, "meta", "cert.pem"))
		if err != nil {
			cleanupPFS(directory)
			fmt.Println("Failed to copy cert file")
			log.Fatalf("Failed to copy cert file: %v", err)
		}
		err = os.Link(key, path.Join(directory, "meta", "key.pem"))
		if err != nil {
			cleanupPFS(directory)
			fmt.Println("Failed to copy key file")
			log.Fatalf("Failed to copy key file: %v", err)
		}
	} else {
		log.Info("Generating certificate.")
		fmt.Println("Generating TLS certificate. Please follow the given instructions.")
		err = tls.GenCertificate(directory)
		if err != nil {
			cleanupPFS(directory)
			fmt.Println("FATAL: Failed to generate certificate")
			log.Fatalf("Failed to generate TLS certificate: %v", err)
		}
	}
}
