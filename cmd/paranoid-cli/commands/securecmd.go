package commands

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/cpssd-students/paranoid/cmd/paranoid-cli/tls"
)

// Secure subcommand
func Secure(c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		_ = cli.ShowCommandHelp(c, "secure")
		os.Exit(0)
	}

	pfsname := args[0]
	usr, err := user.Current()
	if err != nil {
		fmt.Println("FATAL: Couldn't get current user")
		log.Fatalf("cannot get current user: %v", err)
	}

	homeDir := usr.HomeDir
	pfsDir, err := filepath.Abs(path.Join(homeDir, ".pfs", "filesystems", pfsname))
	if err != nil {
		fmt.Println("FATAL: Could not get absolute path to paranoid filesystem.")
		log.Fatalf("could not get absolute path to paranoid filesystem: %v", err)
	}
	if !pathExists(pfsDir) {
		fmt.Println("FATAL: Paranoid filesystem does not exist:", pfsname)
		log.Fatalf("github.com/cpssd-students/paranoid filesystem %s does not exist", pfsname)
	}

	certPath := path.Join(pfsDir, "meta", "cert.pem")
	keyPath := path.Join(pfsDir, "meta", "key.pem")
	if c.Bool("force") {
		os.Remove(certPath)
		os.Remove(keyPath)
	} else {
		if pathExists(certPath) || pathExists(keyPath) {
			fmt.Println("FATAL: Paranoid filesystem already secured.",
				"Run with --force to overwrite existing security files.")
			log.Fatal("Paranoid filesystem already secured.",
				"Run with --force to overwrite existing security files.")
		}
	}

	if (c.String("cert") != "") && (c.String("key") != "") {
		log.Println("Using existing certificate.")
		err = os.Link(c.String("cert"), certPath)
		if err != nil {
			fmt.Println("FATAL: Failed to copy cert file:", err)
			log.Fatalf("failed to copy cert file: %v", err)
		}
		err = os.Link(c.String("key"), keyPath)
		if err != nil {
			fmt.Println("FATAL: Failed to copy key file:", err)
			log.Fatalf("Failed to copy key file: %v", err)
		}
	} else {
		fmt.Println("Generating TLS certificate. Please follow the given instructions.")
		err = tls.GenCertificate(pfsDir)
		if err != nil {
			fmt.Println("FATAL: Failed to generate certificate:", err)
			log.Fatalf("Failed to generate certificate: %v", err)
		}
	}
}
