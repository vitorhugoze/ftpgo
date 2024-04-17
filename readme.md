<h1 align="center">ftpgo</h1>

```
go get github.com/vitorhugoze/ftpgo
```

**ftpgo** Is a ftp implementation developed on top of the default go tcp package made for for allowing file transfer between client and server

**Example usage:**

Start the fpt server:
```go
ftpgo.NewTcpFileServer("localhost:5055").Listen()
```

Now on the client side send a file to the server:
```go
client := ftpgo.NewTcpDataClient("localhost:5055")

if err := client.SendData("<client_source_path>", "<server_destination_path>"); err != nil {
	log.Fatal(err)
}
```

Or request a file from server:
```go
client := ftpgo.NewTcpDataClient("localhost:5055")

if err := client.RequestData("<server_source_path>", "<client_destination_path>"); err != nil {
	log.Fatal(err)
}
```

Note:
You can either send/request a single file or the whole folder from the server
```go
client.RequestData("C:\\source\\file.txt", "C:\\dest\\");
//or
client.RequestData("C:\\source\\", "C:\\dest\\");
```