package pfi

import (
	"log"
	"os"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/pkg/libpfs"
	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

//ParanoidFileSystem is the struct which holds all
//the custom filesystem functions cpssd-students provides
type ParanoidFileSystem struct {
	pathfs.FileSystem
}

//GetAttr is called by fuse when the attributes of a
//file or directory are needed. (pfs stat)
func (fs *ParanoidFileSystem) GetAttr(
	name string, context *fuse.Context,
) (*fuse.Attr, fuse.Status) {
	log.Printf("GetAttr called on %s", name)

	// Special case : "" is the root of our filesystem
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}

	code, stats, err := libpfs.StatCommand(globals.ParanoidDir, name)
	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running stat command: %v", err)
	}

	if err != nil {
		log.Printf("Error running stat command: %v", err)
	}

	if code != returncodes.OK {
		return nil, GetFuseReturnCode(code)
	}

	attr := fuse.Attr{
		Size:  uint64(stats.Length),
		Atime: uint64(stats.Atime.Unix()),
		Ctime: uint64(stats.Ctime.Unix()),
		Mtime: uint64(stats.Mtime.Unix()),
		Mode:  uint32(stats.Mode),
	}

	return &attr, fuse.OK
}

//OpenDir is called when the contents of a directory are needed.
func (fs *ParanoidFileSystem) OpenDir(
	name string, context *fuse.Context,
) ([]fuse.DirEntry, fuse.Status) {
	log.Printf("OpenDir called on %s", name)

	code, fileNames, err := libpfs.ReadDirCommand(globals.ParanoidDir, name)
	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running readdir command: %v", err)
	}

	if err != nil {
		log.Printf("Error running readdir command: %v", err)
	}

	if code != returncodes.OK {
		return nil, GetFuseReturnCode(code)
	}

	dirEntries := make([]fuse.DirEntry, len(fileNames))
	for i, dirName := range fileNames {
		log.Printf("OpenDir has %s", dirName)
		dirEntries[i] = fuse.DirEntry{Name: dirName}
	}

	return dirEntries, fuse.OK
}

//Open is called to get a custom file object for a certain file so that
//Read and Write (among others) opperations can be executed on this
//custom file object (ParanoidFile, see below)
func (fs *ParanoidFileSystem) Open(
	name string, flags uint32, context *fuse.Context,
) (nodefs.File, fuse.Status) {
	log.Printf("Open called on %s ", name)
	return newParanoidFile(name), fuse.OK
}

//Create is called when a new file is to be created.
func (fs *ParanoidFileSystem) Create(
	name string, flags uint32, mode uint32, context *fuse.Context,
) (nodefs.File, fuse.Status) {
	log.Printf("Create called on %s ", name)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestCreatCommand(name, mode)
	} else {
		code, err = libpfs.CreatCommand(globals.ParanoidDir, name, os.FileMode(mode))
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running creat command: %v", err)
	}

	if err != nil {
		log.Printf("Error running creat command: %v", err)
	}

	if code != returncodes.OK {
		return nil, GetFuseReturnCode(code)
	}
	return newParanoidFile(name), fuse.OK
}

//Access is called by fuse to see if it has access to a certain file
func (fs *ParanoidFileSystem) Access(name string, mode uint32, context *fuse.Context) fuse.Status {
	log.Printf("Access called on %s", name)
	if name != "" {
		code, err := libpfs.AccessCommand(globals.ParanoidDir, name, mode)
		if code == returncodes.EUNEXPECTED {
			log.Fatalf("Error running access command: %v", err)
		}

		if err != nil {
			log.Printf("Error running access command: %v", err)
		}
		return GetFuseReturnCode(code)
	}
	return fuse.OK
}

//Rename is called when renaming a file
func (fs *ParanoidFileSystem) Rename(
	oldName string, newName string, context *fuse.Context,
) fuse.Status {
	log.Printf("Rename called on %s to be renamed to %s", oldName, newName)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestRenameCommand(oldName, newName)
	} else {
		code, err = libpfs.RenameCommand(globals.ParanoidDir, oldName, newName)
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running rename command: %v", err)
	}

	if err != nil {
		log.Printf("Error running rename command: %v", err)
	}
	return GetFuseReturnCode(code)
}

//Link creates a hard link from newName to oldName
func (fs *ParanoidFileSystem) Link(
	oldName string, newName string, context *fuse.Context,
) fuse.Status {
	log.Print("Link called")
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestLinkCommand(oldName, newName)
	} else {
		code, err = libpfs.LinkCommand(globals.ParanoidDir, oldName, newName)
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running link command: %v", err)
	}

	if err != nil {
		log.Printf("Error running link command: %s", err)
	}
	return GetFuseReturnCode(code)
}

//Symlink creates a symbolic link from newName to oldName
func (fs *ParanoidFileSystem) Symlink(
	oldName string, newName string, context *fuse.Context,
) fuse.Status {
	log.Printf("Symlink called from %s to %s", oldName, newName)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestSymlinkCommand(oldName, newName)
	} else {
		code, err = libpfs.SymlinkCommand(globals.ParanoidDir, oldName, newName)
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running symlink command: %v", err)
	}

	if err != nil {
		log.Printf("Error running symlink command: %v", err)
	}
	return GetFuseReturnCode(code)
}

// Readlink to where the file is pointing to
func (fs *ParanoidFileSystem) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	log.Printf("Readlink called on %s", name)
	code, link, err := libpfs.ReadlinkCommand(globals.ParanoidDir, name)
	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running readlink command: %v", err)
	}

	if err != nil {
		log.Printf("Error running readlink command: %v", err)
	}
	return link, GetFuseReturnCode(code)
}

//Unlink is called when deleting a file
func (fs *ParanoidFileSystem) Unlink(name string, context *fuse.Context) fuse.Status {
	log.Printf("Unlink called on %s", name)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestUnlinkCommand(name)
	} else {
		code, err = libpfs.UnlinkCommand(globals.ParanoidDir, name)
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running unlink command: %v", err)
	}

	if err != nil {
		log.Printf("Error running unlink command: %v", err)
	}
	return GetFuseReturnCode(code)
}

//Mkdir is called when creating a directory
func (fs *ParanoidFileSystem) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	log.Printf("Mkdir called on %s", name)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestMkdirCommand(name, mode)
	} else {
		code, err = libpfs.MkdirCommand(globals.ParanoidDir, name, os.FileMode(mode))
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running mkdir command: %v", err)
	}

	if err != nil {
		log.Printf("Error running mkdir command: %v", err)
	}
	return GetFuseReturnCode(code)
}

//Rmdir is called when deleting a directory
func (fs *ParanoidFileSystem) Rmdir(name string, context *fuse.Context) fuse.Status {
	log.Printf("Rmdir called on %s", name)
	var code returncodes.Code
	var err error
	if SendOverNetwork {
		code, err = globals.RaftNetworkServer.RequestRmdirCommand(name)
	} else {
		code, err = libpfs.RmdirCommand(globals.ParanoidDir, name)
	}

	if code == returncodes.EUNEXPECTED {
		log.Fatalf("Error running rmdir command: %v", err)
	}

	if err != nil {
		log.Printf("Error running rmdir command: %v", err)
	}
	return GetFuseReturnCode(code)
}

//Truncate is called when a file is to be reduced in length to size.
func (fs *ParanoidFileSystem) Truncate(
	name string, size uint64, context *fuse.Context,
) fuse.Status {
	log.Printf("Truncate called on %s ", name)
	return newParanoidFile(name).Truncate(size)
}

//Utimens update the Access time and modified time of a given file.
func (fs *ParanoidFileSystem) Utimens(
	name string, atime *time.Time, mtime *time.Time, context *fuse.Context,
) fuse.Status {
	log.Printf("Utimens called on %s ", name)
	return newParanoidFile(name).Utimens(atime, mtime)
}

//Chmod is called when the permissions of a file are to be changed
func (fs *ParanoidFileSystem) Chmod(
	name string, perms uint32, context *fuse.Context,
) fuse.Status {
	log.Printf("Chmod called on %s ", name)
	return newParanoidFile(name).Chmod(perms)
}
