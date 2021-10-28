package room

import (
	"net/http"
	"p2p-server/pkg/server"
	"p2p-server/pkg/util"
	"strings"
)

const (
	JoinRoom       = "joinRoom"
	Offer          = "offer"
	Answer         = "answer"
	Candidate      = "candidate"
	HangUp         = "hangUp"
	LeaveRoom      = "leaveRoom"
	UpdateUserList = "updateUserList"
)

type RoomManager struct {
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	var roomManager = &RoomManager{
		rooms: make(map[string]*Room),
	}
	return roomManager
}

type Room struct {
	//User
	users map[string]User
	//Session
	sessions map[string]Session
	ID       string
}

//Room Instance
func NewRoom(id string) *Room {
	var room = &Room{
		users:    make(map[string]User),
		sessions: make(map[string]Session),
		ID:       id,
	}
	return room
}

//GetRoom
func (roomManager *RoomManager) getRoom(id string) *Room {
	return roomManager.rooms[id]
}

//Create Room
func (roomManager *RoomManager) createRoom(id string) *Room {
	roomManager.rooms[id] = NewRoom(id)
	return roomManager.rooms[id]
}

//Delete Room
func (roomManager *RoomManager) deleteRoom(id string) {
	delete(roomManager.rooms, id)
}

//WebSocket Handler
func (roomManager *RoomManager) HandleNewWebSocket(conn *server.WebSocketConn, request *http.Request) {
	util.Infof("On Open %v", request)
	//On message handler
	conn.On("message", func(message []byte) {

		request, err := util.Unmarshal(string(message))
		//error
		if err != nil {
			util.Errorf("Jason Unmarshal error %v", err)
			return
		}

		var data map[string]interface{} = nil
		tmp, found := request["data"]
		if !found {
			util.Errorf("No data found!")
			return
		}
		data = tmp.(map[string]interface{})

		roomId := data["roomId"].(string)
		util.Infof("RoomId: %v", roomId)

		room := roomManager.getRoom(roomId)
		//No ID, create Room
		if room == nil {
			room = roomManager.createRoom(roomId)
		}

		switch request["type"] {
		case JoinRoom:
			onJoinRoom(conn, data, room, roomManager)
			break
		//offer
		case Offer:
			//
			fallthrough
		//Answer
		case Answer:
			//
			fallthrough
		//Candidate
		case Candidate:
			onCandidate(conn, data, room, roomManager, request)
			break
		//HangUp
		case HangUp:
			onHangUp(conn, data, room, roomManager, request)
			break
		default:
			{
				util.Warnf("Unknown request %v", request)
			}
			break
		}
	})

	//On close Handler
	conn.On("close", func(code int, text string) {
		onClose(conn, roomManager)
	})
}

func onJoinRoom(conn *server.WebSocketConn, data map[string]interface{}, room *Room, roomManager *RoomManager) {
	//Create User
	user := User{
		conn: conn,
		info: UserInfo{
			ID:   data["id"].(string),
			Name: data["name"].(string),
		},
	}
	room.users[user.info.ID] = user
	//update User notify
	roomManager.notifyUsersUpdate(conn, room.users)
}

//offer/answer/candidate消息处理
func onCandidate(conn *server.WebSocketConn, data map[string]interface{}, room *Room, roomManager *RoomManager, request map[string]interface{}) {
	//
	to := data["to"].(string)
	//
	if user, ok := room.users[to]; !ok {
		util.Errorf("User Not found[" + to + "]")
		return
	} else {
		//
		user.conn.Send(util.Marshal(request))
	}
}

func onHangUp(conn *server.WebSocketConn, data map[string]interface{}, room *Room, roomManager *RoomManager, request map[string]interface{}) {
	sessionID := data["sessionId"].(string)
	ids := strings.Split(sessionID, "-")

	//
	if user, ok := room.users[ids[0]]; !ok {
		util.Warnf("User [" + ids[0] + "] NotFound")
		return
	} else {
		//
		hangUp := map[string]interface{}{
			//
			"type": HangUp,
			//
			"data": map[string]interface{}{
				//0表示自己 1表示对方
				"to": ids[0],
				//会话Id
				"sessionId": sessionID,
			},
		}
		//发送信息给目标User,即自己[0]
		user.conn.Send(util.Marshal(hangUp))
	}

	//
	if user, ok := room.users[ids[1]]; !ok {
		util.Warnf("User [" + ids[1] + "] Not Found")
		return
	} else {
		//
		hangUp := map[string]interface{}{
			//
			"type": HangUp,
			//
			"data": map[string]interface{}{
				//0表示自己  1表示对方
				"to": ids[1],
				//
				"sessionId": sessionID,
			},
		}
		//发送信息给目标User,即对方[1]
		user.conn.Send(util.Marshal(hangUp))
	}
}

func onClose(conn *server.WebSocketConn, roomManager *RoomManager) {
	util.Infof("Close connection %v", conn)
	var userId string = ""
	var roomId string = ""

	//remove user in the room
	for _, room := range roomManager.rooms {
		for _, user := range room.users {
			//check if user is in connection
			if user.conn == conn {
				userId = user.info.ID
				roomId = room.ID
				break
			}
		}
	}

	if roomId == "" {
		util.Errorf("No such roomId")
		return
	}

	util.Infof("Close roomId %v userId %v", roomId, userId)

	for _, user := range roomManager.getRoom(roomId).users {
		//
		if user.conn != conn {
			leave := map[string]interface{}{
				"type": LeaveRoom,
				"data": userId,
			}
			user.conn.Send(util.Marshal(leave))
		}
	}
	util.Infof("delete user", userId)
	//
	delete(roomManager.getRoom(roomId).users, userId)

	roomManager.notifyUsersUpdate(conn, roomManager.getRoom(roomId).users)
}

//Notifu User Update
func (roomManager *RoomManager) notifyUsersUpdate(conn *server.WebSocketConn, users map[string]User) {
	//
	infos := []UserInfo{}
	//
	for _, userClient := range users {
		//
		infos = append(infos, userClient.info)
	}
	//
	request := make(map[string]interface{})
	//
	request["type"] = UpdateUserList
	//
	request["data"] = infos
	//
	for _, user := range users {
		//
		user.conn.Send(util.Marshal(request))
	}
}
