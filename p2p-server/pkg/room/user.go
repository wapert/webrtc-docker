package room

import (
	"p2p-server/pkg/server"
)

//UserInfo
type UserInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

//User
type User struct {
	//UserInfo
	info UserInfo
	//ws Conn
	conn *server.WebSocketConn
}

//Session
type Session struct {
	//session ID
	id string
	//From
	from User
	//To
	to User
}
