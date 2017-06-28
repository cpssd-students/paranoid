// +build !integration

package commands

import (
	"math/rand"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/pp2p/paranoid/libpfs/encryption"
	"github.com/pp2p/paranoid/libpfs/returncodes"
	log "github.com/pp2p/paranoid/logger"
)

var testDirectory string

func TestMain(m *testing.M) {
	defer removeTestDir()

	log.Info("Running tests with no encryption")
	encryption.Encrypted = false
	noEncryption := m.Run()
	if noEncryption != 0 {
		os.Exit(noEncryption)
	}

	log.Info("Running tests with encryption")
	encryption.Encrypted = true
	encryptionKey := []byte("86F7E437FAA5A7FCE15D1DDCB9EAEAEA")
	cipherBlock, err := encryption.GenerateAESCipherBlock(encryptionKey)
	if err != nil {
		log.Fatal("Could not create cipherBlock", err)
	}
	encryption.SetCipher(cipherBlock)
	os.Exit(m.Run())
}

func createTestDir() {
	if err := os.RemoveAll(testDirectory); err != nil {
		log.Fatalf("error creating test directory %s: %v", testDirectory, err)
	}

	if err := os.Mkdir(testDirectory, 0777); err != nil {
		log.Fatalf("error creating test directory %s: %v", testDirectory, err)
	}
}

func removeTestDir() {
	os.RemoveAll(testDirectory)
}

func setupTestDirectory() {
	createTestDir()

	if code, err := InitCommand(testDirectory); code != returncodes.OK {
		log.Fatal("error initing directory for testing:", err)
	}
}

func TestSimpleCommandUsage(t *testing.T) {
	setupTestDirectory()

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Errorf("error creating test file: %v", err)
	}

	code, _, err = WriteCommand(testDirectory, "test.txt", -1, -1, []byte("BLAH #1"))
	if code != returncodes.OK {
		t.Errorf("Write did not return OK, got %s: %v", code.String(), err)
	}

	code, returnData, err := ReadCommand(testDirectory, "test.txt", -1, -1)
	if code != returncodes.OK {
		t.Errorf("Read did not return OK, got %s: %v", code.String(), err)
	}

	if string(returnData) != "BLAH #1" {
		t.Errorf("Output does not match BLAH #1, got %s", string(returnData))
	}
}

func TestComplexCommandUsage(t *testing.T) {
	setupTestDirectory()

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Errorf("error creating test file: %v", err)
	}

	code, bytesWritten, err := WriteCommand(testDirectory, "test.txt", -1, -1, []byte("START"))
	if code != returncodes.OK {
		t.Errorf("Write did not return OK, got %s: %v", code.String(), err)
	}
	if bytesWritten != len([]byte("START")) {
		t.Errorf("Wrote %d bytes, wanted %d", bytesWritten, len([]byte("START")))
	}

	code, returnData, err := ReadCommand(testDirectory, "test.txt", 2, 2)
	if code != returncodes.OK {
		t.Errorf("Read did not return OK, got %s: %v", code.String(), err)
	}

	if string(returnData) != "AR" {
		t.Errorf("expected AR from partial read, got %s", string(returnData))
	}

	code, bytesWritten, err = WriteCommand(testDirectory, "test.txt", 5, -1, []byte("END"))
	if code != returncodes.OK {
		t.Errorf("Write did not return OK, got %s: %v", code.String(), err)
	}
	if bytesWritten != len([]byte("END")) {
		t.Errorf("Write %d bytes, wanted %d", bytesWritten, len([]byte("END")))
	}

	code, returnData, err = ReadCommand(testDirectory, "test.txt", -1, -1)
	if code != returncodes.OK {
		t.Error("Read did not return OK. Actual:", code, " Error:", err)
	}

	if string(returnData) != "STARTEND" {
		t.Errorf("Full read does not match STARTEND, got %s", string(returnData))
	}

	code, files, err := ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Errorf("Read did not return OK, got %s: %v", code.String(), err)
	}

	if files[0] != "test.txt" || len(files) > 1 {
		t.Errorf("Readdir got incorrect result: %v", files)
	}
}

func TestFilePermissionsCommands(t *testing.T) {
	setupTestDirectory()

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("error creating test file:", err)
	}

	code, statIn, err := StatCommand(testDirectory, "test.txt")
	if code != returncodes.OK {
		t.Error("Stat did not return OK. Actual:", code, " Error:", err)
	}

	if statIn.Mode.Perm() != 0777 {
		t.Error("Incorrect file permissions = ", statIn.Mode.Perm())
	}

	code, err = ChmodCommand(testDirectory, "test.txt", os.FileMode(0377))
	if code != returncodes.OK {
		t.Error("error chmoding file:", err)
	}

	code, statIn, err = StatCommand(testDirectory, "test.txt")
	if code != returncodes.OK {
		t.Error("Stat did not return OK. Actual:", code, " Error:", err)
	}

	if statIn.Mode.Perm() != 0377 {
		t.Error("Incorrect file permissions = ", statIn.Mode.Perm())
	}

	code, err = ChmodCommand(testDirectory, "test.txt", os.FileMode(0500))
	if code != returncodes.OK {
		t.Error("error chmoding file:", err)
	}

	code, statIn, err = StatCommand(testDirectory, "test.txt")
	if code != returncodes.OK {
		t.Error("Stat did not return OK. Actual:", code, " Error:", err)
	}

	if statIn.Mode.Perm() != 0500 {
		t.Error("Incorrect file permissions = ", statIn.Mode.Perm())
	}

	code, err = AccessCommand(testDirectory, "test.txt", 4)
	if code != returncodes.OK {
		t.Error("Access command did not return OK. Actual:", code, " Error:", err)
	}

	code, err = AccessCommand(testDirectory, "test.txt", 2)
	if code != returncodes.EACCES {
		t.Errorf("Access command did not return EACCES, got %s: %v", code.String(), err)
	}
}

func TestENOENT(t *testing.T) {
	setupTestDirectory()

	code, _, _ := ReadCommand(testDirectory, "test.txt", -1, -1)
	if code != returncodes.ENOENT {
		t.Error("Read did not return ENOENT. Actual:", code)
	}

	code, _, _ = StatCommand(testDirectory, "test.txt")
	if code != returncodes.ENOENT {
		t.Error("Stat did not return ENOENT. Actual:", code)
	}

	code, _, _ = WriteCommand(testDirectory, "test.txt", -1, -1, []byte("data"))
	if code != returncodes.ENOENT {
		t.Error("Write did not return ENOENT. Actual:", code)
	}
}

func TestFilesystemCommands(t *testing.T) {
	setupTestDirectory()

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("creat did not return OK. Error:", err)
	}

	code, files, err := ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if files[0] != "test.txt" || len(files) > 1 {
		t.Error("Readdir got incorrect result")
	}

	code, err = RenameCommand(testDirectory, "test.txt", "test2.txt")
	if code != returncodes.OK {
		t.Error("error renaming test.txt:", err)
	}

	code, files, err = ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if files[0] != "test2.txt" || len(files) > 1 {
		t.Error("Readdir got incorrect result")
	}

	code, err = UnlinkCommand(testDirectory, "test2.txt")
	if code != returncodes.OK {
		t.Error("Unlink command did not return OK. Actual:", code, " Error:", err)
	}

	code, files, err = ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) > 0 {
		t.Error("Readdir got incorrect result")
	}
}

func TestLinkCommand(t *testing.T) {
	setupTestDirectory()

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("creat did not return OK. Error:", err)
	}

	code, err = LinkCommand(testDirectory, "test.txt", "test2.txt")
	if code != returncodes.OK {
		t.Error("link command did not return OK. Actual:", code, " Error:", err)
	}

	code, files, err := ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if files[0] != "test.txt" && files[1] != "test.txt" {
		t.Error("Readdir got incorrect result")
	}

	if files[0] != "test2.txt" && files[1] != "test2.txt" {
		t.Error("Readdir got incorrect result")
	}

	if len(files) != 2 {
		t.Error("Readdir got incorrect results")
	}

	code, bytesWritten, err := WriteCommand(testDirectory, "test2.txt", -1, -1, []byte("hellotest"))
	if code != returncodes.OK {
		t.Error("Write did not return OK. Actual:", code, " Error:", err)
	}
	if bytesWritten != len([]byte("hellotest")) {
		t.Error("Write did not return correct number of bytes Actual:", bytesWritten, "Expected:", len([]byte("hellotest")))
	}

	code, data, err := ReadCommand(testDirectory, "test.txt", -1, -1)
	if code != returncodes.OK {
		t.Error("Read did not return OK. Actual:", code, " Error :", err)
	}

	if string(data) != "hellotest" {
		t.Error("Read did not return correct result. Actual: ", string(data))
	}

	code, err = UnlinkCommand(testDirectory, "test.txt")
	if code != returncodes.OK {
		t.Error("Unlink did not return OK. Actual:", code, " Error:", err)
	}

	code, files, err = ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) != 1 {
		t.Error("Readdir got incorrect result")
	}

	if files[0] != "test2.txt" {
		t.Error("Readdir got incorrect result")
	}

	code, data, err = ReadCommand(testDirectory, "test2.txt", -1, -1)
	if code != returncodes.OK {
		t.Error("Read did not return OK. Actual:", code, " Error:", err)
	}

	if string(data) != "hellotest" {
		t.Error("Read did not return correct result. Actual : ", string(data))
	}
}

func TestSymlinkcommand(t *testing.T) {
	setupTestDirectory()
	code, err := SymlinkCommand(testDirectory, "testfolder/testlink", "testsymlink")
	if code != returncodes.OK {
		t.Error("Symlink did not return OK. Actual:", code, " Error:", err)
	}

	code, linkConents, err := ReadlinkCommand(testDirectory, "testsymlink")
	if code != returncodes.OK {
		t.Error("Readlink did not return OK. Actual:", code, " Error:", err)
	}
	if linkConents != "testfolder/testlink" {
		t.Error("Unexpected link contents : ", linkConents)
	}

	code, stats, err := StatCommand(testDirectory, "testsymlink")
	if code != returncodes.OK {
		t.Error("Statcommand did not return OK. Actual:", code, " Error:", err)
	}
	if int(stats.Mode)&int(syscall.S_IFLNK) == 0 {
		t.Error("Does not appear as a symlink from stat:", stats.Mode)
	}

	code, _, err = ReadCommand(testDirectory, "testsymlink", -1, -1)
	if code != returncodes.EIO {
		t.Error("Should return EIO when attempting to read a symlink. Actual :", code, " Error:", err)
	}
}

func TestUtimes(t *testing.T) {
	setupTestDirectory()

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("error creating test file:", err)
	}

	atime := time.Unix(100, 100)
	mtime := time.Unix(500, 250)
	code, err = UtimesCommand(testDirectory, "test.txt", &atime, &mtime)
	if code != returncodes.OK {
		t.Error("Utimes did not return OK. Actual:", code, " Error:", err)
	}

	code, statIn, err := StatCommand(testDirectory, "test.txt")
	if code != returncodes.OK {
		t.Error("Stat did not return OK. Actual:", code, " Error:", err)
	}

	if statIn.Atime != time.Unix(100, 100) {
		t.Error("Incorrect stat time. Acutal: ", statIn.Atime)
	}

	if statIn.Mtime != time.Unix(500, 250) {
		t.Error("Incorrect stat time. Acutal: ", statIn.Atime)
	}
}

func TestTruncate(t *testing.T) {
	setupTestDirectory()

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("error creating test file:", err)
	}

	code, bytesWritten, err := WriteCommand(testDirectory, "test.txt", -1, -1, []byte("HI!!!!!"))
	if code != returncodes.OK {
		t.Error("Write command failed! : ", err)
	}
	if bytesWritten != len([]byte("HI!!!!!")) {
		t.Error("Write did not return correct number of bytes Actual:", bytesWritten, "Expected:", len([]byte("HI!!!!!")))
	}

	code, err = TruncateCommand(testDirectory, "test.txt", 3)
	if code != returncodes.OK {
		t.Error("truncate did not return OK. Actual:", code, " Error:", err)
	}

	code, data, err := ReadCommand(testDirectory, "test.txt", -1, -1)
	if code != returncodes.OK {
		t.Error("Read command did not return OK. Actual:", code, " Error:", err)
	}

	if string(data) != "HI!" {
		t.Error("Read command returned incorrect output ", string(data))
	}
}

func TestSimpleDirectoryUsage(t *testing.T) {
	setupTestDirectory()

	code, err := MkdirCommand(testDirectory, "documents", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("Mkdir did not return OK. Actual:", code, " Error:", err)
	}

	code, files, err := ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) != 1 {
		t.Error("Readdir returned something other than one file: ", files)
	}

	if files[0] != "documents" {
		t.Error("File is not equal to 'documents':", files[0])
	}

	code, err = RmdirCommand(testDirectory, "documents")
	if code != returncodes.OK {
		t.Error("rmdir did not return OK. Actual:", code, " Error:", err)
	}

	code, files, err = ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) != 0 {
		t.Error("Readdir returned more than 0: ", files)
	}
}

func TestComplexDirectoryUsage(t *testing.T) {
	setupTestDirectory()

	// directory within directory
	code, err := MkdirCommand(testDirectory, "documents", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("Mkdir did not return OK. Actual:", code, " Error:", err)
	}

	code, _ = MkdirCommand(testDirectory, "documents/work_docs", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("Mkdir did not return OK. Actual:", code, " Error:", err)
	}

	code, files, err := ReadDirCommand(testDirectory, "documents")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) != 1 {
		t.Error("Readdir returned something other than one file: ", files)
	}

	if files[0] != "work_docs" {
		t.Error("File is not equal to 'work_docs':", files[0])
	}

	// file within directory
	code, err = CreatCommand(testDirectory, "documents/important_links.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("Mkdir did not return OK. Actual:", code, " Error:", err)
	}

	code, files, err = ReadDirCommand(testDirectory, "documents")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) != 2 {
		t.Error("Readdir returned something other than 2 files: ", files)
	}

	if (files[0] != "important_links.txt" && files[1] != "work_docs") && (files[1] != "important_links.txt" && files[0] != "work_docs") {
		t.Error("File is not equal to 'important_links.txt':", files[0])
	}

	// writing and reading from file within directory
	toWrite := []byte("https://www.google.com/")
	code, bytesWritten, err := WriteCommand(testDirectory, "documents/important_links.txt", -1, -1, toWrite)
	if code != returncodes.OK {
		t.Error("Write did not return OK. Actual:", code, " Error:", err)
	}
	if bytesWritten != len(toWrite) {
		t.Error("Write did not return correct number of bytes Actual:", bytesWritten, "Expected:", len(toWrite))
	}

	code, data, err := ReadCommand(testDirectory, "documents/important_links.txt", -1, -1)
	if code != returncodes.OK {
		t.Error("Read did not return OK. Actual:", code, " Error:", err)
	}

	if string(data) != "https://www.google.com/" {
		t.Error("Read did not return 'https://www.google.com/', Actual:", string(data))
	}

	// link files in different directories
	code, err = LinkCommand(testDirectory, "documents/important_links.txt", "documents/work_docs/worklinks.txt")
	if code != returncodes.OK {
		t.Error("Link did not return OK. Actual:", code, " Error:", err)
	}

	code, data, err = ReadCommand(testDirectory, "documents/work_docs/worklinks.txt", -1, -1)
	if code != returncodes.OK {
		t.Error("Read did not return OK. Actual:", code, " Error:", err)
	}

	if string(data) != "https://www.google.com/" {
		t.Error("Read did not return 'https://www.google.com/', Actual:", string(data))
	}

	// remove directory with contents inside
	code, err = RmdirCommand(testDirectory, "documents/work_docs")
	if code == returncodes.OK {
		t.Error("Rmdir returned ok when it should have returned ENOTEMPTY")
	}

	code, err = UnlinkCommand(testDirectory, "documents/work_docs/worklinks.txt")
	if code != returncodes.OK {
		t.Error("Unlink failed to unlink: ", code, " Error:", err)
	}

	code, err = RmdirCommand(testDirectory, "documents/work_docs")
	if code != returncodes.OK {
		t.Error("Rmdir failed on empty directory:", code, " Error:", err)
	}

	code, files, err = ReadDirCommand(testDirectory, "documents")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) != 1 {
		t.Error("Readdir returned something other than 1 file: ", files)
	}

	if files[0] != "important_links.txt" {
		t.Error("File is not equal to 'important_links.txt':", files[0])
	}

	// writing and reading from a directory
	code, _, err = WriteCommand(testDirectory, "documents", -1, -1, []byte("Should not work"))
	if code == returncodes.OK {
		t.Error("Succeeded to write to a directory")
	}

	code, _, err = ReadCommand(testDirectory, "documents", -1, -1)
	if code == returncodes.OK {
		t.Error("Succeeded to read from a directory")
	}

	// renaming a directory
	code, err = RenameCommand(testDirectory, "documents", "docs")
	if code != returncodes.OK {
		t.Error("Rename failed on a directory:", code, " Error:", err)
	}

	code, files, err = ReadDirCommand(testDirectory, "")
	if code != returncodes.OK {
		t.Error("Readdir did not return OK. Actual:", code, " Error:", err)
	}

	if len(files) != 1 {
		t.Error("Readdir returned something other than 1 file: ", files)
	}

	if files[0] != "docs" {
		t.Error("File is not equal to 'docs':", files[0])
	}
}

func randN(x int) int {
	if x == 0 {
		return 0
	}
	return rand.Intn(x)
}

func TestComplexReadWrite(t *testing.T) {
	setupTestDirectory()

	seed := time.Now().UnixNano()
	log.Info("Test seed: ", seed)
	rand.Seed(seed)

	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		t.Error("error creating test file:", err)
	}

	currentFileData := make([]byte, 1024)
	fileLength := 0

	//Perform 100 random writes
	for i := 0; i < 100; i++ {
		maxWriteStart := fileLength
		if maxWriteStart > 900 {
			maxWriteStart = 900
		}
		writeStart := randN(maxWriteStart)
		writeLength := randN(50) + 10
		data := make([]byte, writeLength)
		for j := 0; j < writeLength; j++ {
			data[j] = byte(randN(26) + int('A'))
		}

		code, bytesWritten, err := WriteCommand(testDirectory, "test.txt", int64(writeStart), int64(writeLength), data)
		if code != returncodes.OK {
			t.Error("Write did not return OK. Actual:", code, " Error:", err)
		}
		if bytesWritten != len(data) {
			t.Error("Write did not return correct number of bytes Actual:", bytesWritten, "Expected:", len(data))
		}

		if writeStart+writeLength > fileLength {
			fileLength = writeStart + writeLength
		}

		copy(currentFileData[writeStart:writeStart+writeLength], data)
	}

	code, returnData, err := ReadCommand(testDirectory, "test.txt", -1, -1)
	if code != returncodes.OK {
		t.Error("Read did not return OK. Actual:", code, " Error:", err)
	}

	if string(returnData) != string(currentFileData[:fileLength]) {
		t.Errorf("Output does not match \n Expected: %s\n Actual: %s\n", string(currentFileData[:fileLength]), string(returnData))
	}
}
