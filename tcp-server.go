package ftpgo

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"path/filepath"
)

type TcpFileServer struct {
	ServerAddress string
	//Set if server will listen for multiple connections or just one and then close
	Persistant bool
}

func NewTcpFileServer(serverAddress string, persistant ...bool) TcpFileServer {

	var persistantAux bool

	if len(persistant) > 0 {
		persistantAux = persistant[0]
	} else {
		persistantAux = true
	}

	return TcpFileServer{
		ServerAddress: serverAddress,
		Persistant:    persistantAux,
	}
}

func (server TcpFileServer) Listen() {

	lis, err := net.Listen("tcp", server.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	if server.Persistant {

		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Println(err)
			}

			go dataReceiver(conn)
		}
	} else {

		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
		}

		dataReceiver(conn)
	}

}

func dataReceiver(conn net.Conn) {

	var err error
	var fileName string
	var singleFile bool

	defer conn.Close()

	if err = binary.Read(conn, binary.BigEndian, &singleFile); err != nil {
		log.Fatal(err)
	}

	destPath, err := readFileInfo(conn)
	if err != nil {
		log.Fatal(err)
	}

	if singleFile {

		fileName, err = readFileInfo(conn)
		if err != nil {
			log.Fatal(err)
		}

		err = fileReader(conn, destPath, fileName)
	} else {
		err = folderReader(conn, destPath)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func folderReader(conn net.Conn, destPath string) error {

	for {

		var endOfTransmission bool

		err := binary.Read(conn, binary.BigEndian, &endOfTransmission)
		if err != nil {
			return err
		}

		if endOfTransmission {
			break
		}

		fileName, err := readFileInfo(conn)
		if err != nil {
			return err
		}

		err = fileReader(conn, destPath, fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
Reads file data from connection and copy it into destpath
*/
func fileReader(conn net.Conn, destpath, fileName string) error {

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
func readFileInfo(conn net.Conn) (string, error) {

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
