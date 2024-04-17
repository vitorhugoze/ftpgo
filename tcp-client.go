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
	BufSize       int
	ServerAddress string
}

/*
Creates a new handler that can be used to send and request data from server
*/
func NewTcpDataClient(serverAddress string) TcpDataClient {

	return TcpDataClient{
		BufSize:       16384,
		ServerAddress: serverAddress,
	}
}

func (dataClient TcpDataClient) WithBufferSize(BufSize int) TcpDataClient {
	dataClient.BufSize = BufSize
	return dataClient
}

/*
Send file or folder from LocalPath on the client to ServerPath on server
*/
func (dataClient TcpDataClient) SendData(sourcePath, destPath string) error {

	var singleFile bool

	buf := new(bytes.Buffer)
	dataChan := make(chan []byte)

	if filepath.Ext(sourcePath) == "" {
		singleFile = false
	} else {
		singleFile = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tcp.TcpDataWriter(dataChan, dataClient.ServerAddress, ctx)

	//Tell server if is sending or requesting data
	if err := binary.Write(buf, binary.BigEndian, true); err != nil {
		return err
	}

	//Tell server if is sending a single file or a folder
	if err := binary.Write(buf, binary.BigEndian, singleFile); err != nil {
		return err
	}
	dataChan <- buf.Bytes()

	//Tell server the destination path
	if err := tcp.WriteFileInfo(dataChan, []byte(filepath.Dir(destPath))); err != nil {
		return err
	}

	//Check if is a folder or just a file
	if singleFile {

		f, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer f.Close()

		//Writes filename to the server
		if err = tcp.WriteFileInfo(dataChan, []byte(filepath.Base(sourcePath))); err != nil {
			return err
		}

		return tcp.FileWriter(dataChan, dataClient.BufSize, f)
	} else {
		return tcp.FolderWriter(dataChan, dataClient.BufSize, sourcePath)
	}
}

func (dataClient TcpDataClient) RequestData(sourcePath, destPath string) error {

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
	if err := tcp.WriteFileInfo(dataChan, []byte(sourcePath)); err != nil {
		return err
	}

	//Check if is folder or file
	if filepath.Ext(sourcePath) == "" {
		err = tcp.FolderReader(conn, filepath.Dir(destPath))
	} else {

		if filepath.Ext(destPath) == "" {
			err = tcp.FileReader(conn, filepath.Dir(destPath), filepath.Base(sourcePath))
		} else {
			err = tcp.FileReader(conn, filepath.Dir(destPath), filepath.Base(destPath))
		}
	}

	return err
}
