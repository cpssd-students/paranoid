package libpfsbenchmark

import (
	"log"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	. "github.com/cpssd-students/paranoid/pkg/libpfs"
	"github.com/cpssd-students/paranoid/pkg/libpfs/returncodes"
)

var testDirectory string

func TestMain(m *testing.M) {
	testDirectory = path.Join(os.TempDir(), "paranoidTest")
	defer removeTestDir()
	os.Exit(m.Run())
}

func removeTestDir() {
	os.RemoveAll(testDirectory)
}

func createTestDir() {
	err := os.RemoveAll(testDirectory)
	if err != nil {
		log.Fatalf("error creating test directory: %v", err)
	}

	err = os.Mkdir(testDirectory, 0777)
	if err != nil {
		log.Fatalf("error creating test directory: %v", err)
	}
}

func setupTestDirectory() {
	createTestDir()

	code, err := InitCommand(testDirectory)
	if code != returncodes.OK {
		log.Fatalf("error initing directory for testing: %v", err)
	}
}

func BenchmarkCreat(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := CreatCommand(testDirectory, "test.txt"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error creating test file: %v", err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := CreatCommand(testDirectory, "test.txt"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error creating test file: %v", err)
		}
		code, _, err = WriteCommand(testDirectory, "test.txt"+str, 0, 0, []byte("Hello World"))
		if code != returncodes.OK {
			log.Fatalf("error writing to test file: %v", err)
		}
	}
}

func BenchmarkRename(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := CreatCommand(testDirectory, "test.txt"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error creating test file: %v", err)
		}
		code, err = RenameCommand(testDirectory, "test.txt"+str, "test2.txt"+str)
		if code != returncodes.OK {
			log.Fatalf("error renaming test file: %v", err)
		}
	}
}

func BenchmarkRead(b *testing.B) {
	setupTestDirectory()
	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		log.Fatalf("error creating test file: %v", err)
	}
	for n := 0; n < b.N; n++ {
		code, err, _ := ReadCommand(testDirectory, "test.txt", 0, 0)
		if code != returncodes.OK {
			log.Fatalf("error reading test file: %v", err)
		}
	}
}

func BenchmarkStat(b *testing.B) {
	setupTestDirectory()
	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		log.Fatalf("Error creating test file: %v", err)
	}
	for n := 0; n < b.N; n++ {
		code, _, err := StatCommand(testDirectory, "test.txt")
		if code != returncodes.OK {
			log.Fatalf("error stat test file: %v", err)
		}
	}
}

func BenchmarkAccess(b *testing.B) {
	setupTestDirectory()
	code, err := CreatCommand(testDirectory, "test.txt", os.FileMode(0777))
	if code != returncodes.OK {
		log.Fatal("error creating test file:", err)
	}
	for n := 0; n < b.N; n++ {
		code, err := AccessCommand(testDirectory, "test.txt", 0)
		if code != returncodes.OK {
			log.Fatal("error accessing test file:", err)
		}
	}
}

func BenchmarkTruncate(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := CreatCommand(testDirectory, "test.txt"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error creating test file: %v", err)
		}
		code, err = TruncateCommand(testDirectory, "test.txt"+str, 3)
		if code != returncodes.OK {
			log.Fatalf("error truncating test file: %v", err)
		}
	}
}

func BenchmarkUtimes(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		atime := time.Unix(100, 100)
		mtime := time.Unix(500, 250)
		code, err := CreatCommand(testDirectory, "test.txt"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error creating test file: %v", err)
		}
		code, err = UtimesCommand(testDirectory, "test.txt"+str, &atime, &mtime)
		if code != returncodes.OK {
			log.Fatalf("error changing test file time: %v", err)
		}
	}
}

func BenchmarkRmDir(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := MkdirCommand(testDirectory, "testDir"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error making benchdir: %v", err)
		}
		code, err = RmdirCommand(testDirectory, "testDir"+str)
		if code != returncodes.OK {
			log.Fatalf("error removing benchdir: %v", err)
		}
	}
}

func BenchmarkReadDir(b *testing.B) {
	setupTestDirectory()
	code, err := MkdirCommand(testDirectory, "testDir", os.FileMode(0777))
	if code != returncodes.OK {
		log.Fatalf("error making benchDir: %v", err)
	}
	code, err = CreatCommand(testDirectory, path.Join("testDir", "test.txt"), os.FileMode(0777))
	if code != returncodes.OK {
		log.Fatalf("error creating test file: %v", err)
	}
	for n := 0; n < b.N; n++ {
		code, _, err = ReadDirCommand(testDirectory, "testDir")
		if code != returncodes.OK {
			log.Fatalf("error reading benchDir: %v", err)
		}
	}
}

func BenchmarkLink(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := CreatCommand(testDirectory, "test.txt"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error creating testFile: %v", err)
		}
		_, _ = LinkCommand(testDirectory, "test.txt"+str, "test2.txt"+str)
	}
}

func BenchmarkSymLink(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := SymlinkCommand(testDirectory, "testfolder/testlink", "testsymlink"+str)
		if code != returncodes.OK {
			log.Fatalf("Symlink did not return OK. Actual %v: %v", code, err)
		}
	}
}

func BenchmarkReadLink(b *testing.B) {
	setupTestDirectory()
	code, err := SymlinkCommand(testDirectory, "testfolder/testlink", "testsymlink")
	if code != returncodes.OK {
		log.Fatalf("Symlink did not return OK. Actual: %v: %v", code, err)
	}
	for n := 0; n < b.N; n++ {
		code, _, err = ReadlinkCommand(testDirectory, "testsymlink")
		if code != returncodes.OK {
			log.Printf("Readlink did not return OK. Actual: %v: %v", code, err)
		}
	}
}

func BenchmarkMkDir(b *testing.B) {
	setupTestDirectory()
	for n := 0; n < b.N; n++ {
		str := strconv.Itoa(n)
		code, err := MkdirCommand(testDirectory, "testDir"+str, os.FileMode(0777))
		if code != returncodes.OK {
			log.Fatalf("error making benchdir: %v", err)
		}
	}
}
