package server

import (
	"errors"
	"net"
	"p2p-server/pkg/util"
	"sync"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
)

//HeartBeat interval 5 sec
const pingPeriod = 5 * time.Second

type WebSocketConn struct {
	//emitter
	emission.Emitter
	//ws socker
	socket *websocket.Conn
	mutex  *sync.Mutex
	closed bool
}

func NewWebSocketConn(socket *websocket.Conn) *WebSocketConn {
	var conn WebSocketConn
	conn.Emitter = *emission.NewEmitter()
	conn.socket = socket
	conn.mutex = new(sync.Mutex)
	conn.closed = false
	conn.socket.SetCloseHandler(func(code int, text string) error {
		util.Warnf("%s [%d]", text, code)
		//send close state
		conn.Emit("close", code, text)
		conn.closed = true
		return nil
	})
	return &conn
}

//read WS msg
func (conn *WebSocketConn) ReadMessage() {
	in := make(chan []byte)
	stop := make(chan struct{})
	pingTicker := time.NewTicker(pingPeriod)

	var c = conn.socket
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				util.Warnf("ws Read Error: %v", err)
				if c, k := err.(*websocket.CloseError); k {
					conn.Emit("close", c.Code, c.Text)
				} else {
					if c, k := err.(*net.OpError); k {
						conn.Emit("close", 1008, c.Error())
					}
				}
				close(stop)
				break
			}
			in <- message
		}
	}()
	for {
		select {
		case _ = <-pingTicker.C:
			util.Infof("Send HeartBeat...")
			//send empty packet
			heartPackage := map[string]interface{}{
				"type": "heartPackage",
				"data": "",
			}
			//send HB to peer
			if err := conn.Send(util.Marshal(heartPackage)); err != nil {
				util.Errorf("Send Heart Beat Error")
				pingTicker.Stop()
				return
			}
		case message := <-in:
			{
				util.Infof("Receive Msg: %s", message)
				conn.Emit("message", []byte(message))
			}
		case <-stop:
			return
		}
	}
}

func (conn *WebSocketConn) Send(message string) error {
	util.Infof("Send data: %s", message)
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	if conn.closed {
		return errors.New("websocket: write closed")
	}
	return conn.socket.WriteMessage(websocket.TextMessage, []byte(message))
}

//close WS
func (conn *WebSocketConn) Close() {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	if conn.closed == false {
		util.Infof("Close WS socket : ", conn)
		conn.socket.Close()
		conn.closed = true
	} else {
		util.Warnf("WS socket is closed :", conn)
	}
}
