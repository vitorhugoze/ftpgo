package tcp

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
)

func DataResponder(conn net.Conn) error {

	var bufSize int64

	dataChan := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go TcpResponseWriter(dataChan, conn, ctx)

	if err := binary.Read(conn, binary.BigEndian, &bufSize); err != nil {
		return err
	}

	path, err := ReadFileInfo(conn)
	if err != nil {
		return err
	}

	//Check if is folder or file
	if filepath.Ext(path) == "" {

		if err = FolderWriter(dataChan, int(bufSize), path); err != nil {
			log.Fatal(err)
		}
	} else {

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		if err = FileWriter(dataChan, int(bufSize), f); err != nil {
			return err
		}
	}

	return nil
}

/*
Write data from all the content on the sourcePath to dataChan
*/
func FolderWriter(dataChan chan<- []byte, bufSize int, sourcePath string) error {

	var currPath string

	err := filepath.Walk(sourcePath, func(path string, info fs.FileInfo, err error) error {

		buf := new(bytes.Buffer)

		if err != nil {
			return err
		}

		if info.IsDir() {

			currPath, err = filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}

		} else {

			//Tell server that this is not end of transmission
			if err := binary.Write(buf, binary.BigEndian, false); err != nil {
				return err
			}
			dataChan <- buf.Bytes()

			//Writes file relative path and name to server
			if err = WriteFileInfo(dataChan, []byte(filepath.Join(currPath, filepath.Base(path)))); err != nil {
				return err
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if err := FileWriter(dataChan, bufSize, f, info); err != nil {
				return err
			}
		}

		return nil
	})

	buf := new(bytes.Buffer)

	//Tell server that this is the end of transmission
	if err := binary.Write(buf, binary.BigEndian, true); err != nil {
		return err
	}

	dataChan <- buf.Bytes()

	return err
}

/*
Write file data to dataChan
*/
func FileWriter(dataChan chan<- []byte, bufSize int, f *os.File, fInfo ...fs.FileInfo) error {

	var err error
	var fInfoAux fs.FileInfo

	readenBytes := 0

	if len(fInfo) == 0 {
		fInfoAux, err = f.Stat()

		if err != nil {
			return err
		}
	} else {
		fInfoAux = fInfo[0]
	}

	for {

		buf := new(bytes.Buffer)

		data := make([]byte, bufSize)

		n, err := io.ReadFull(f, data)
		if err != nil && n == 0 {
			return err
		}
		readenBytes += n

		//Tell server how many bytes are being sent
		err = binary.Write(buf, binary.BigEndian, int64(n))
		if err != nil {
			return err
		}

		//Check if is last chunk and add this information to the buffer
		if n < bufSize || readenBytes == int(fInfoAux.Size()) {
			if err = binary.Write(buf, binary.BigEndian, true); err != nil {
				return err
			}
		} else {
			if err = binary.Write(buf, binary.BigEndian, false); err != nil {
				return err
			}
		}

		_, err = buf.Write(data[:n])
		if err != nil {
			return err
		}

		dataChan <- buf.Bytes()

		if n < bufSize || readenBytes == int(fInfoAux.Size()) {
			break
		}

	}

	return err
}

/*
Writes information like file name and destination path to dataChan
*/
func WriteFileInfo(dataChan chan<- []byte, fileInfo []byte) error {

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, int64(len(fileInfo))); err != nil {
		return err
	}

	if _, err := buf.Write(fileInfo); err != nil {
		return err
	}

	dataChan <- buf.Bytes()

	return nil
}

/*
Forwards data from dataChan to the tcp server
*/
func TcpDataWriter(dataChan <-chan []byte, serverAddress string, ctx context.Context) {

	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	for {

		select {
		case <-ctx.Done():
			return
		case c := <-dataChan:
			_, err = conn.Write(c)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

}

/*
Forwards data from dataChan to client as a response
*/
func TcpResponseWriter(dataChan <-chan []byte, conn net.Conn, ctx context.Context) {

	for {

		select {
		case <-ctx.Done():
			return
		case c := <-dataChan:
			_, err := conn.Write(c)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

}
