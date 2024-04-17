package ftpgo

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/vitorhugoze/ftpgo/pkg/tcp"
)

type TcpDataClient struct {
	BufSize                             int
	SingleFile                          bool
	ServerAddress, SourcePath, DestPath string
}

/*
Creates a new handler that can be used to send and request data from server
*/
func NewTcpDataClient(serverAddress, sourcePath, destPath string) TcpDataClient {

	var singleFile bool

	if filepath.Ext(sourcePath) == "" {
		singleFile = false
	} else {
		singleFile = true
	}

	return TcpDataClient{
		BufSize:       16384,
		SingleFile:    singleFile,
		ServerAddress: serverAddress,
		SourcePath:    sourcePath,
		DestPath:      destPath,
	}
}

func (dataClient TcpDataClient) WithBufferSize(BufSize int) TcpDataClient {
	dataClient.BufSize = BufSize
	return dataClient
}

/*
Send file or folder from LocalPath on the client to ServerPath on server
*/
func (dataClient TcpDataClient) SendData() error {

	buf := new(bytes.Buffer)
	dataChan := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tcp.TcpDataWriter(dataChan, dataClient.ServerAddress, ctx)

	//Tell server if is sending or requesting data
	if err := binary.Write(buf, binary.BigEndian, true); err != nil {
		return err
	}

	//Tell server if is sending a single file or a folder
	if err := binary.Write(buf, binary.BigEndian, dataClient.SingleFile); err != nil {
		return err
	}
	dataChan <- buf.Bytes()

	//Tell server the destination path
	if err := tcp.WriteFileInfo(dataChan, []byte(filepath.Dir(dataClient.DestPath))); err != nil {
		return err
	}

	//Check if is a folder or just a file
	if dataClient.SingleFile {

		f, err := os.Open(dataClient.SourcePath)
		if err != nil {
			return err
		}
		defer f.Close()

		//Writes filename to the server
		if err = tcp.WriteFileInfo(dataChan, []byte(filepath.Base(dataClient.SourcePath))); err != nil {
			return err
		}

		return tcp.FileWriter(dataChan, dataClient.BufSize, f)
	} else {
		return tcp.FolderWriter(dataChan, dataClient.BufSize, dataClient.SourcePath)
	}
}

func (dataClient TcpDataClient) RequestData() error {

	buf := new(bytes.Buffer)
	dataChan := make(chan []byte)

	conn, err := net.Dial("tcp", dataClient.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tcp.TcpResponseWriter(dataChan, conn, ctx)

	//Tell server if is sending or requesting data
	if err := binary.Write(buf, binary.BigEndian, false); err != nil {
		return err
	}

	if err := binary.Write(buf, binary.BigEndian, int64(dataClient.BufSize)); err != nil {
		return err
	}
	dataChan <- buf.Bytes()

	//Writes file path to server
	if err := tcp.WriteFileInfo(dataChan, []byte(dataClient.SourcePath)); err != nil {
		return err
	}

	//Check if is folder or file
	if filepath.Ext(dataClient.SourcePath) == "" {
		err = tcp.FolderReader(conn, filepath.Dir(dataClient.DestPath))
	} else {

		if filepath.Ext(dataClient.DestPath) == "" {
			err = tcp.FileReader(conn, filepath.Dir(dataClient.DestPath), filepath.Base(dataClient.SourcePath))
		} else {
			err = tcp.FileReader(conn, filepath.Dir(dataClient.DestPath), filepath.Base(dataClient.DestPath))
		}
	}

	return err
}
