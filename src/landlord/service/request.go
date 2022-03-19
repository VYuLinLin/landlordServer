package service

import (
	"sort"
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
			keys := []int{}
			for k := range room.Tables {
				keys = append(keys, int(k))
			}
			sort.Ints(keys)
			for _, k := range keys {
				otherTable := room.Tables[TableId(k)]
				if len(otherTable.TableClients) < 3 {
					table = otherTable
					break
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
		if (client.Table == nil) {
			var data = make(map[string]UserId)
			data["userId"] = client.UserInfo.UserId
			client.sendMsg(RoomLeave, 200, data)
			return
		}
		if client.Status >= Calling || client.Table.State >= GamePushCard {
			client.sendMsg(RoomLeave, 500, "游戏中不能离开房间！")
			return
		}
		client.Status = INVALID
		var userId = client.UserInfo.UserId
		var table = client.Table
		var tableId TableId
		if client.Table != nil {
			tableId = client.Table.TableId
			client.Table.leaveTable(client)
		}
		if client.Room != nil {
			client.Room.leaveRoom(client, tableId)
		}
		defer func() {
			var data = make(map[string]UserId)
			data["userId"] = userId
			client.sendMsg(RoomLeave, 200, data)
			for _, c := range table.TableClients {
				data["creatorId"] = table.Creator.UserInfo.UserId
				c.sendMsg(RoomLeave, 200, data)
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
		var data = make(map[UserId]bool)
		for _, c := range client.Table.TableClients {
			data[c.UserInfo.UserId] = c.Ready
		}
		for _, c := range client.Table.TableClients {
			c.sendMsg(PlayerReady, 200, data)
		}
		go client.Table.gameStart()

	case PlayerCallPoints:
		data := r.Data
		var score int
		if id, ok := data.(string); ok {
			score, _ = strconv.Atoi(id)
		}

		//client.IsCalled = true
		client.Table.callPoints(client.UserInfo.UserId, score)

	}
}
