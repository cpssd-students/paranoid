package returncodes

// Code is the basic type for all returncodes
type Code int

// Current libpfs supported return codes
const (
	OK          Code = iota
	ENOENT           //No such file or directory.
	EACCES           //Can not access file
	EEXIST           //File already exists
	ENOTEMPTY        //Directory not empty
	EISDIR           //Is Directory
	EIO              //Input/Output error
	ENOTDIR          //Isn't Directory
	EBUSY            //System is busy
	EUNEXPECTED      //Unforseen error
)

var codes = [...]string{
	"OK",
	"ENOENT",
	"EACCES",
	"EEXIST",
	"ENOTEMPTY",
	"EISDIR",
	"EIO",
	"ENOTDIR",
	"EBUSY",
	"EUNEXPECTED",
}

// Strings returns a human readable representation of the codes
func (c Code) String() string { return codes[c] }
