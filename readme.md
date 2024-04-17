<h1 align="center">ftpgo</h1>

**ftpgo** Is a ftp implementation developed on top of the default go tcp package made for for allowing file transfer between client and server

**Example usage:**

Start the fpt server:
```go
ftpgo.NewTcpFileServer("localhost:5055").Listen()
```

Now on the client side send a file to the server:
```go
client := ftpgo.NewTcpDataClient("localhost:5055", "<client_source_path>", "<server_destination_path>")

if err := client.SendData(); err != nil {
	log.Fatal(err)
}
```

Or request a file from server:
```go
client := ftpgo.NewTcpDataClient("localhost:5055", "<server_source_path>", "<client_destination_path>")

if err := client.RequestData(); err != nil {
	log.Fatal(err)
}
```