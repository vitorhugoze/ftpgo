package tcp

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
)

/*
Check if data is folder or single file and calls apropriate function
*/
func DataReader(conn net.Conn) error {

	var err error
	var fileName string
	var singleFile bool

	defer conn.Close()

	if err = binary.Read(conn, binary.BigEndian, &singleFile); err != nil {
		return err
	}

	destPath, err := ReadFileInfo(conn)
	if err != nil {
		return err
	}

	if singleFile {

		fileName, err = ReadFileInfo(conn)
		if err != nil {
			return err
		}

		err = FileReader(conn, destPath, fileName)
	} else {
		err = FolderReader(conn, destPath)
	}

	if err != nil {
		return err
	}

	return nil
}

/*
Reads folder data from connection and copy it into destPath
*/
func FolderReader(conn net.Conn, destPath string) error {

	for {

		var endOfTransmission bool

		err := binary.Read(conn, binary.BigEndian, &endOfTransmission)
		if err != nil {
			return err
		}

		if endOfTransmission {
			break
		}

		fileName, err := ReadFileInfo(conn)
		if err != nil {
			return err
		}

		err = FileReader(conn, destPath, fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
Reads file data from connection and copy it into destpath
*/
func FileReader(conn net.Conn, destpath, fileName string) error {

	var dataSize int64
	var endOfTransmission bool

	fileDir := filepath.Dir(filepath.Join(destpath, fileName))

	err := os.MkdirAll(fileDir, fs.ModeDir)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(destpath, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	for {

		err = binary.Read(conn, binary.BigEndian, &dataSize)
		if err != nil {
			return err
		}

		err = binary.Read(conn, binary.BigEndian, &endOfTransmission)
		if err != nil {
			return err
		}

		_, err = io.CopyN(f, conn, dataSize)

		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if endOfTransmission {
			break
		}

	}

	return nil
}

/*
Reads file information from the connection
*/
func ReadFileInfo(conn net.Conn) (string, error) {

	var err error
	var fileInfoLen int64

	buf := new(bytes.Buffer)

	if err = binary.Read(conn, binary.BigEndian, &fileInfoLen); err != nil {
		return "", err
	}

	if _, err = io.CopyN(buf, conn, fileInfoLen); err != nil {
		return "", err
	}

	return buf.String(), nil
}
