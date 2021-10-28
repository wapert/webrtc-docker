package main

import (
	"os"
	"p2p-server/pkg/room"
	"p2p-server/pkg/server"
	"p2p-server/pkg/util"

	"gopkg.in/ini.v1"
)

//program entry
func main() {
	//load config
	cfg, err := ini.Load("configs/config.ini")
	if err != nil {
		util.Errorf("Load config Error: %v", err)
		os.Exit(1)
	}

	//room Manager Init
	roomManager := room.NewRoomManager()
	//P2p Server Init
	wsServer := server.NewP2PServer(roomManager.HandleNewWebSocket)

	//Read ssl Cert
	sslCert := cfg.Section("general").Key("cert").String()
	//Read ssl Key
	sslKey := cfg.Section("general").Key("key").String()
	//Bind IP
	bindAddress := cfg.Section("general").Key("bind").String()

	//Bind Port
	port, err := cfg.Section("general").Key("port").Int()
	//Default port
	if err != nil {
		port = 8000
	}
	//set html root
	htmlRoot := cfg.Section("general").Key("html_root").String()

	//set p2p default config
	config := server.DefaultConfig()
	config.Host = bindAddress
	config.Port = port
	config.CertFile = sslCert
	config.KeyFile = sslKey
	config.HTMLRoot = htmlRoot

	wsServer.Bind(config)
}
