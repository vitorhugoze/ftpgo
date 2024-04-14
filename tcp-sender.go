package ftpgo

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
)

type TcpFileSender struct {
	BufSize                             int
	SingleFile                          bool
	ServerAddress, SourcePath, DestPath string
}

func NewTcpFileSender(serverAddress, sourcePath, destPath string) TcpFileSender {

	var singleFile bool

	if filepath.Ext(sourcePath) == "" {
		if filepath.Ext(destPath) != "" {
			log.Fatal(errors.New("can't export from a local forder to a remote file, both need to be folders"))
		}

		singleFile = false
	} else {
		singleFile = true
	}

	return TcpFileSender{
		BufSize:       16384,
		SingleFile:    singleFile,
		ServerAddress: serverAddress,
		SourcePath:    sourcePath,
		DestPath:      destPath,
	}
}

func (fileSender TcpFileSender) WithBufferSize(BufSize int) TcpFileSender {
	fileSender.BufSize = BufSize
	return fileSender
}

func (fileSender TcpFileSender) SendData() error {

	buf := new(bytes.Buffer)
	dataChan := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go fileSender.tcpDataSender(dataChan, ctx)

	//Tell server if is sending a single file or a folder
	if err := binary.Write(buf, binary.BigEndian, fileSender.SingleFile); err != nil {
		return err
	}
	dataChan <- buf.Bytes()

	//Tell server the destination path
	if err := writeFileInfo(dataChan, []byte(filepath.Dir(fileSender.DestPath))); err != nil {
		return err
	}

	//Check if is a folder or just a file
	if fileSender.SingleFile {

		f, err := os.Open(fileSender.SourcePath)
		if err != nil {
			return err
		}
		defer f.Close()

		//Writes filename to the server
		if err = writeFileInfo(dataChan, []byte(filepath.Base(fileSender.SourcePath))); err != nil {
			return err
		}

		return fileSender.fileWriter(dataChan, f)
	} else {
		return fileSender.folderWriter(dataChan)
	}
}

/*
Write data from all the content on the sourcePath to dataChan
*/
func (fileSender TcpFileSender) folderWriter(dataChan chan<- []byte) error {

	var currPath string

	err := filepath.Walk(fileSender.SourcePath, func(path string, info fs.FileInfo, err error) error {

		buf := new(bytes.Buffer)

		if err != nil {
			return err
		}

		if info.IsDir() {

			currPath, err = filepath.Rel(fileSender.SourcePath, path)
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
			if err = writeFileInfo(dataChan, []byte(filepath.Join(currPath, filepath.Base(path)))); err != nil {
				return err
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if err := fileSender.fileWriter(dataChan, f, info); err != nil {
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
func (fileSender TcpFileSender) fileWriter(dataChan chan<- []byte, f *os.File, fInfo ...fs.FileInfo) error {

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

		data := make([]byte, fileSender.BufSize)

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
		if n < fileSender.BufSize || readenBytes == int(fInfoAux.Size()) {
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

		if n < fileSender.BufSize || readenBytes == int(fInfoAux.Size()) {
			break
		}

	}

	return err
}

/*
Writes information like file name and destination path to dataChan
*/
func writeFileInfo(dataChan chan<- []byte, fileInfo []byte) error {

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
func (fileSender TcpFileSender) tcpDataSender(dataChan <-chan []byte, ctx context.Context) {

	conn, err := net.Dial("tcp", fileSender.ServerAddress)
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
