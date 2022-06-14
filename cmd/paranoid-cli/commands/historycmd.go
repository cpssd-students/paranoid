package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"

	"github.com/urfave/cli"

	"github.com/cpssd-students/paranoid/pkg/raft"
	pb "github.com/cpssd-students/paranoid/proto/raft"
)

// History subcommand
func History(c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		_ = cli.ShowCommandHelp(c, "history")
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	target := args[0]
	if fileSystemExists(target) {
		target = path.Join(usr.HomeDir, ".pfs", "filesystems", target, "meta", "raft", "raft_logs")
	}
	read(target, c)
}

// read shows the history of a log in the given directory
func read(directory string, c *cli.Context) {
	filePath := path.Join(os.TempDir(), "log.pfslog")
	logsToLogfile(directory, filePath, c)
	defer os.Remove(filePath)
	less := exec.Command("less", filePath)
	less.Stdout = os.Stdout
	less.Stdin = os.Stdin
	less.Stderr = os.Stderr
	_ = less.Run()
}

// logsToLogFile converts the binary logs in the logDir paramenter
// to a human readable file in the given filePath paramenter.
func logsToLogfile(logDir, filePath string, c *cli.Context) {
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		log.Printf("failed reading directory %s: %v", logDir, err)
		_ = cli.ShowCommandHelp(c, "history")
		os.Exit(1)
	}
	writeFile, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("unable to create raft log file: %v", err)
	}
	defer writeFile.Close()

	numRunes := len(strconv.Itoa(len(files)))
	numRunesString := strconv.Itoa(numRunes)
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		p, err := fileToProto(file, logDir)
		if err != nil {
			log.Fatalf("unable to parse binary proto: %v", err)
		}

		_, _ = writeFile.WriteString(toLine(i+1, numRunesString, p))
	}
}

// toLine converts a protobuf object to a human readable string representation
// the pad parameter is the number of runes to be used for writing the logNum
// and Term in string form.
func toLine(logNum int, pad string, p *pb.LogEntry) string {
	marker := fmt.Sprintf("%-"+pad+"d Term: %-"+pad+"d", logNum, p.Term)

	if p.Entry.Type == pb.Entry_StateMachineCommand {
		return fmt.Sprintf("%s Command: %s\n",
			marker, commandString(p.Entry.Command))
	} else if p.Entry.Type == pb.Entry_ConfigurationChange {
		return fmt.Sprintf("%s ConfigChange: %s\n",
			marker, configurationString(p.Entry.Config))
	} else {
		return fmt.Sprintf("%s Demo: %v", marker, p.Entry.Demo)
	}
}

func configurationString(conf *pb.Configuration) string {
	typ := fmt.Sprintf("%-21s", configTypeString(conf.Type))
	conf.Type = 0
	return fmt.Sprintf("%s%v", typ, conf)
}

// commandString returns the string representation of a StateMachineCommand
func commandString(cmd *pb.StateMachineCommand) string {
	typeStr := commandTypeString(raft.ActionType(cmd.Type))
	cmd.Type = 0
	size := len(cmd.Data)
	if size > 0 {
		cmd.Data = nil
		return fmt.Sprintf("%-9s %v Data: %s", typeStr, cmd, bytesString(size))
	}
	return fmt.Sprintf("%-9s %v", typeStr, cmd)
}

// bytesString returns the human readable representation of a data size
// TODO: Allow values over 3GB (int32 max)
func bytesString(size int) string {
	p := 0
	for ; size > 1000; size, p = size/1000, p+1 {
	}
	return fmt.Sprintf("%d%s", size, []string{"B", "KB", "MB", "GB", "TB"}[p])
}

// commandTypeString returns the string representation of a command log type.
func commandTypeString(ty raft.ActionType) string {
	switch ty {
	case raft.TypeWrite:
		return "Write"
	case raft.TypeCreat:
		return "Creat"
	case raft.TypeChmod:
		return "Chmod"
	case raft.TypeTruncate:
		return "Truncate"
	case raft.TypeUtimes:
		return "Utimes"
	case raft.TypeRename:
		return "Rename"
	case raft.TypeLink:
		return "Link"
	case raft.TypeSymlink:
		return "Symlink"
	case raft.TypeUnlink:
		return "Unlink"
	case raft.TypeMkdir:
		return "Mkdir"
	case raft.TypeRmdir:
		return "Rmdir"
	default:
		return "Unknown"
	}
}

// configTypeString returns the string representation of a configuration change log type.
func configTypeString(ty pb.Configuration_ConfigurationType) string {
	switch ty {
	case pb.Configuration_CurrentConfiguration:
		return "Current Configuration"
	case pb.Configuration_FutureConfiguration:
		return "Future Configuration"
	default:
		return "Unknown"
	}
}
