package test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/vitorhugoze/ftpgo"
)

func TestSendFileData(t *testing.T) {

	go ftpgo.NewTcpFileServer("<server_address:port>", false).Listen()
	time.Sleep(300)

	err := ftpgo.NewTcpDataClient("<server_address:port>", "<server_address>", "<dest_path>").SendData()
	if err != nil {
		t.Error("Error sending data to server", err)
	}

	fSource, err := os.Open("<server_address:port>")
	if err != nil {
		t.Error("Error opening source file ", err)
	}

	fCopy, err := os.Open("<dest_file>")
	if err != nil {
		t.Error("Error opening copied file", err)
	}

	err = compareFileBytes(fSource, fCopy)
	if err != nil {
		t.Error(err)
	}
}

func TestRequestFileData(t *testing.T) {

	go ftpgo.NewTcpFileServer("<server_address:port>", false).Listen()
	time.Sleep(300)

	err := ftpgo.NewTcpDataClient("<server_address:port>", "<server_address>", "<dest_path>").RequestData()
	if err != nil {
		t.Error("Error sending data to server", err)
	}

	fSource, err := os.Open("<server_address:port>")
	if err != nil {
		t.Error("Error opening source file ", err)
	}

	fCopy, err := os.Open("<dest_file>")
	if err != nil {
		t.Error("Error opening copied file", err)
	}

	err = compareFileBytes(fSource, fCopy)
	if err != nil {
		t.Error(err)
	}
}

func compareFileBytes(f1, f2 *os.File) error {

	var err error
	readen1 := 1024
	readen2 := 1024

	b1 := make([]byte, 1024)
	b2 := make([]byte, 1024)

	for readen1 > 0 && readen2 > 0 {

		readen1, err = f1.Read(b1)
		if err != nil && err != io.EOF {
			return err
		}

		readen2, err = f2.Read(b2)
		if err != nil && err != io.EOF {
			return err
		}

		if readen1 != readen2 || !bytes.Equal(b1, b2) {
			return errors.New("bytes slices not equal")
		}

	}

	return nil
}
