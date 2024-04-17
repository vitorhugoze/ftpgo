package ftpgo

import (
	"encoding/binary"
	"log"
	"net"

	"github.com/vitorhugoze/ftpgo/pkg/tcp"
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

			go sendOrReceiveData(conn)
		}
	} else {

		conn, err := lis.Accept()
		if err != nil {
			log.Fatal(err)
		}

		sendOrReceiveData(conn)
	}
}

func sendOrReceiveData(conn net.Conn) {

	var err error
	var receiveData bool

	if err := binary.Read(conn, binary.BigEndian, &receiveData); err != nil {
		log.Fatal(err)
	}

	//Check if is sending or requesting data
	if receiveData {
		err = tcp.DataReader(conn)
	} else {
		err = tcp.DataResponder(conn)
	}

	if err != nil {
		log.Fatal(err)
	}
}
