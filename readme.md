<h1 align="center">ftpgo</h1>

**ftpgo** Is a ftp implementation developed on top of the default go tcp package made for for allowing file transfer between client and server

**Example usage:**

Start the fpt server:
```go
ftpgo.NewTcpFileServer("localhost:5055").Listen()
```

Now on the client side send a file to the server:
```go
sender := ftpgo.NewTcpFileSender("localhost:5055", "<source_file>", "<server_destination_path>")

if err := sender.SendData(); err != nil {
	log.Fatal(err)
}
```