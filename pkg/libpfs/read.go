package libpfs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"paranoid/pkg/libpfs/encryption"
	"paranoid/pkg/libpfs/returncodes"
	log "paranoid/pkg/logger"
)

//ReadCommand reads data from a file
func ReadCommand(paranoidDirectory, filePath string, offset, length int64) (returnCode returncodes.Code, fileContents []byte, returnError error) {
	log.V(1).Infof("read called on %s in %s", filePath, paranoidDirectory)

	namepath := getParanoidPath(paranoidDirectory, filePath)

	err := GetFileSystemLock(paranoidDirectory, SharedLock)
	if err != nil {
		return returncodes.EUNEXPECTED, nil, err
	}

	defer func() {
		err := UnLockFileSystem(paranoidDirectory)
		if err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
			fileContents = nil
		}
	}()

	fileType, err := getFileType(paranoidDirectory, namepath)
	if err != nil {
		return returncodes.EUNEXPECTED, nil, err
	}

	if fileType == typeENOENT {
		return returncodes.ENOENT, nil, fmt.Errorf("%s does not exist", filePath)
	}

	if fileType == typeDir {
		return returncodes.EISDIR, nil, fmt.Errorf("%s is a paranoidDirectory", filePath)
	}

	if fileType == typeSymlink {
		return returncodes.EIO, nil, fmt.Errorf("%s is a symlink", filePath)
	}

	inodeBytes, code, err := getFileInode(namepath)
	if code != returncodes.OK || err != nil {
		return code, nil, err
	}
	inodeFileName := string(inodeBytes)

	err = getFileLock(paranoidDirectory, inodeFileName, SharedLock)
	if err != nil {
		return returncodes.EUNEXPECTED, nil, err
	}

	defer func() {
		err := unLockFile(paranoidDirectory, inodeFileName)
		if err != nil {
			returnCode = returncodes.EUNEXPECTED
			returnError = err
			fileContents = nil
		}
	}()

	file, err := os.OpenFile(path.Join(paranoidDirectory, "contents", inodeFileName), os.O_RDONLY, 0777)
	if err != nil {
		return returncodes.EUNEXPECTED, nil, fmt.Errorf("error opening contents file: %s", err)
	}
	defer file.Close()

	var fileBuffer bytes.Buffer
	bytesRead := make([]byte, 1024)

	maxRead := 100000000

	if offset == -1 {
		offset = 0
	}

	if length != -1 {
		maxRead = int(length)
	}

	for {
		n, readerror, err := readAt(file, bytesRead, offset)
		if err != nil {
			return returncodes.EUNEXPECTED, nil, fmt.Errorf("error reading file %s", err)
		}

		if n > maxRead {
			bytesRead = bytesRead[0:maxRead]
			_, err := fileBuffer.Write(bytesRead)
			if err != nil {
				return returncodes.EUNEXPECTED, nil, fmt.Errorf("error writing to file buffer: %s", err)
			}
			break
		}

		offset = offset + int64(n)
		maxRead = maxRead - n
		if readerror == io.EOF {
			bytesRead = bytesRead[:n]
			_, err := fileBuffer.Write(bytesRead)
			if err != nil {
				return returncodes.EUNEXPECTED, nil, fmt.Errorf("error writing to file buffer: %s", err)
			}
			break
		}

		if readerror != nil {
			return returncodes.EUNEXPECTED, nil, fmt.Errorf("error reading from %s: %s", filePath, err)
		}

		bytesRead = bytesRead[:n]
		_, err = fileBuffer.Write(bytesRead)
		if err != nil {
			return returncodes.EUNEXPECTED, nil, fmt.Errorf("error writing to file buffer: %s", err)
		}
	}
	return returncodes.OK, fileBuffer.Bytes(), nil
}

func readAt(file *os.File, bytesRead []byte, offset int64) (n int, readerror error, err error) {
	if !encryption.Encrypted {
		n, readerror := file.ReadAt(bytesRead, offset)
		return n, readerror, nil
	}

	if len(bytesRead) == 0 {
		return 0, nil, nil
	}

	cipherSizeInt64 := int64(encryption.CipherSize())
	extraStartBytes := offset % cipherSizeInt64
	extraEndBytes := cipherSizeInt64 - ((offset + int64(len(bytesRead))) % cipherSizeInt64)
	readStart := 1 + offset - extraStartBytes
	newBytesRead := make([]byte, int64(len(bytesRead))+extraStartBytes+extraEndBytes)

	fileLength, err := getFileLength(file)
	if err != nil {
		return 0, nil, err
	}

	n, readerror = file.ReadAt(newBytesRead, readStart)
	n = n - int(extraStartBytes)
	if n > len(bytesRead) {
		n = len(bytesRead)
	}
	if offset+int64(n) > fileLength {
		n = int(fileLength - offset)
	}

	err = encryption.Decrypt(newBytesRead)
	if err != nil {
		return 0, nil, err
	}
	newBytesRead = newBytesRead[extraStartBytes:]
	copy(bytesRead, newBytesRead)
	return n, readerror, nil
}
