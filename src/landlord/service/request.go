package service

import (
	"strconv"

	"github.com/astaxie/beego/logs"
)

type Request struct {
	Action string `json:"action"`// 请求接口
	Data   interface{} `json:"data"`
}

type Response struct {
	Action string `json:"action"`// 推送接口
	Code   int `json:"code"`
	Data   interface{} `json:"data"`
}

// 处理websocket请求
func wsRequest(r Request, client *Client) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("wsRequest panic:%v ", err)
			client.sendMsg(r.Action, 500, err)
		}
	}()
	switch r.Action {
	case RoomList:
		var rooms []*Room
		for _, room := range roomManager.Rooms {
			rooms = append(rooms, room)
		}
		client.sendMsg(RoomList, 200, rooms)

	case RoomJoinSelf:
		var table = client.Table
		if table != nil {
			table.joinTable(client)
			return
		}
		data := r.Data
		var roomId int
		if id, ok := data.(string); ok {
			roomId, _ = strconv.Atoi(id)
		}
		//roomManager.Lock.RLock()
		//defer roomManager.Lock.RUnlock()
		if room, ok := roomManager.Rooms[roomId]; ok {
			client.Room = room
			for _, otherTable := range client.Room.Tables {
				if len(otherTable.TableClients) < 3 {
					table = otherTable
				}
			}
			if table == nil {
				table = client.Room.newTable(client)
			}
			//client.Room.Lock.RLock()
			//defer client.Room.Lock.RUnlock()

			table.joinTable(client)
			//client.sendMsg([]interface{}{common.ResJoinRoom, res})
		}
	case RoomLeave:
		if client.Status >= Calling || client.Table.State >= GamePushCard {
			client.sendMsg(RoomLeave, 500, "游戏中不能离开房间！")
			return
		}
		client.Status = INVALID
		var tableId TableId
		if client.Table != nil {
			tableId = client.Table.TableId
			client.Table.leaveTable(client)
		}
		if client.Room != nil {
			client.Room.leaveRoom(client, tableId)
		}
		defer func() {
			client.sendMsg(RoomLeave, 200, client.UserInfo.UserId)
			if client.Table == nil {
				return
			}
			for _, c := range client.Table.TableClients {
				c.sendMsg(RoomLeave, 200, client.UserInfo.UserId)
			}
		}()
	case TableInfo:
		data := r.Data.(map[string]interface{})
		var roomId float64
		var tableId float64
		if id, ok := data["room_id"]; ok {
			roomId = id.(float64)
		}
		if id, ok := data["table_id"]; ok {
			tableId = id.(float64)
		}
		if room, ok := roomManager.Rooms[int(roomId)]; ok {
			if table, ok := room.Tables[TableId(tableId)]; ok {
				table.getTableData(client)
			}
		}

	case PlayerReady:
		client.Ready = true
		client.sendMsg(PlayerReady, 200, nil)
		go client.Table.gameStart()

	case PlayerCallPoints:
		data := r.Data
		var score int
		if id, ok := data.(string); ok {
			score, _ = strconv.Atoi(id)
		}
		if score > client.Table.GameManage.MaxCallScore {
			client.Table.GameManage.MaxCallScore = score
			client.Table.GameManage.MaxCallScoreTurn = client
		}
		//client.IsCalled = true
		client.Table.callPoints(client.UserInfo.UserId, score)

	}
}
