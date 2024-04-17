package test

import (
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/vitorhugoze/ftpgo"
)

/*
Runs benchmark with multiple buffer sizes for the same file
*/
func BenchmarkSenderBufferSize(b *testing.B) {

	go ftpgo.NewTcpFileServer("<server_adress:port>").Listen()

	time.Sleep(300 * time.Millisecond)

	for j := range 10 {

		bufSize := math.Pow(2, float64(j+10))

		b.Run(fmt.Sprint("Running test with buffer size: ", bufSize), func(b *testing.B) {

			sender := ftpgo.NewTcpDataClient("<server_adress:port>", "<source_file>", "<dest_path>")
			sender.WithBufferSize(int(bufSize))

			if err := sender.SendData(); err != nil {
				log.Fatal(err)
			}
		})

	}

}
