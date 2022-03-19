package service

import (
	"bytes"
	"encoding/json"
	"landlord/common"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 1 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512

	RoleFarmer   = 0
	RoleLandlord = 1
)
const (
	INVALID = iota // 无效
	UNREADY        // 未准备
	ready          // 已准备
	Calling        // 抢地主（叫分）
	double         // 加倍
	PLAYING        // 出牌阶段
	over           // 游戏结束
)

var (
	newline  = []byte{'\n'}
	space    = []byte{' '}
	upGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Method != "GET" {
				logs.Error("method is not GET")
				return false
			}
			if r.URL.Path != "/ws/" {
				logs.Error("path error")
				return false
			}
			return true
		},
	} //不验证origin
)

type UserId int

type UserInfo struct {
	UserId   UserId
	Username string
	Coin     int
	Role     int
}

type Client struct {
	conn       *websocket.Conn
	UserInfo   *UserInfo
	Room       *Room
	Table      *Table
	HandPokers []int
	Status     int
	Ready      bool
	IsCalled   bool    //是否叫完分
	Next       *Client //链表
	IsRobot    bool
	toRobot    chan []interface{} //发送给robot的消息
	toServer   chan []interface{} //robot发送给服务器
	Mux sync.RWMutex
}

//重置状态
func (c *Client) reset() {
	c.UserInfo.Role = 1
	c.HandPokers = make([]int, 0, 21)
	c.Ready = false
	c.IsCalled = false
}

//发送房间内已有的牌桌信息
func (c *Client) sendRoomTables() {
	//res := make([][2]int, 0)
	//for _, table := range c.Room.Tables {
	//	if len(table.TableClients) < 3 {
	//		res = append(res, [2]int{int(table.TableId), len(table.TableClients)})
	//	}
	//}
	//c.sendMsg([]interface{}{common.ResTableList, res})
}

//func (c *Client) sendMsg(msg []interface{}) {
//	if c.IsRobot {
//		c.toRobot <- msg
//		return
//	}
//	var msgByte []byte
//	var err error
//	if msg[0] == common.ResHeart {
//		heart, _ := strconv.Atoi(common.ResHeart)
//		msgByte, _ = json.Marshal(heart)
//	} else {
//		msgByte, err = json.Marshal(msg)
//		if err != nil {
//			logs.Error("send msg [%v] marsha1 err:%v", string(msgByte), err)
//			return
//		}
//	}
//	err = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
//	if err != nil {
//		logs.Error("send msg SetWriteDeadline [%v] err:%v", string(msgByte), err)
//		return
//	}
//	w, err := c.conn.NextWriter(websocket.BinaryMessage)
//	if err != nil {
//		err = c.conn.Close()
//		if err != nil {
//			logs.Error("close client err: %v", err)
//		}
//	}
//	_, err = w.Write(msgByte)
//	if err != nil {
//		logs.Error("Write msg [%v] err: %v", string(msgByte), err)
//	}
//	if err = w.Close(); err != nil {
//		err = c.conn.Close()
//		if err != nil {
//			logs.Error("close err: %v", err)
//		}
//	}
//}
func (c *Client) sendMsg(action string, code int, data interface{}) {
	res := Response{
		action,
		code,
		data,
	}
	var msgByte []byte
	var err error
	if action == common.ResHeart {
		heart, _ := strconv.Atoi(common.ResHeart)
		msgByte, _ = json.Marshal(heart)
	} else {
		msgByte, err = json.Marshal(res)
		if err != nil {
			logs.Error("send msg [%v] marsha1 err:%v", string(msgByte), err)
			return
		}
	}
	//err = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	//if err != nil {
	//	logs.Error("send msg SetWriteDeadline [%v] err:%v", string(msgByte), err)
	//	return
	//}
	c.Mux.Lock()
	w, err := c.conn.NextWriter(websocket.BinaryMessage)

	if err != nil {
		err = c.conn.Close()
		if err != nil {
			logs.Error("close client err: %v", err)
		}
	}
	_, err = w.Write(msgByte)
	c.Mux.Unlock()

	if err != nil {
		logs.Error("Write msg [%v] err: %v", string(msgByte), err)
	}
	if err = w.Close(); err != nil {
		err = c.conn.Close()
		if err != nil {
			logs.Error("close err: %v", err)
		}
	}
}

//光比客户端
func (c *Client) close() {
	return
	if c.Table != nil {
		for _, client := range c.Table.TableClients {
			if c.Table.Creator == c && c != client {
				c.Table.Creator = client
			}
			if c == client.Next {
				client.Next = nil
			}
		}
		if len(c.Table.TableClients) != 1 {
			for _, client := range c.Table.TableClients {
				if client != client.Table.Creator {
					client.Table.Creator.Next = client
				}
			}
		}
		if len(c.Table.TableClients) == 1 {
			c.Table.Creator = nil
			//delete(c.Room.Tables, c.Table.TableId)
			return
		}
		delete(c.Table.TableClients, c.UserInfo.UserId)
		if c.Table.State == GamePlaying {
			c.Table.syncUser()
			//c.Table.reset()
		}
		if c.IsRobot {
			close(c.toRobot)
			close(c.toServer)
		}
	}
}

//可能是因为版本问题，导致有些未处理的error
func (c *Client) readPump() {
	defer func() {
		logs.Debug("readPump exit")
		//c.conn.Close()
		c.close()
		//if c.Room.AllowRobot {
		//	if c.Table != nil {
		//		for _, client := range c.Table.TableClients {
		//			client.close()
		//		}
		//	}
		//}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	//c.conn.SetReadDeadline(time.Now().Add(pongWait))
	//c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logs.Error("websocket user_id[%d] unexpected close error: %v", c.UserInfo.UserId, err)
			}
			break
		}
		if string(message) == common.ReqHeart {
			c.sendMsg(common.ResHeart, 0, nil)
			continue
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		req := Request{}
		//var data []interface{}
		err = json.Unmarshal(message, &req)
		if err != nil {
			logs.Error("message unmarsha1 err, user_id[%d] err:%v", c.UserInfo.UserId, err)
		} else {
			wsRequest(req, c)
		}
	}
}

// websocket 关闭监听
func (c *Client) setCloseHandler() {
	c.conn.SetCloseHandler(func(code int, text string) error {
		DeleteClient(c.UserInfo.UserId)
		logs.Error("The user [%s] disabled the Websocket: code = %v", c.UserInfo.Username, code)
		return nil
	})
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	userId, err := strconv.Atoi(params.Get("id"))
	userName := params.Get("name")

	if err != nil {
		logs.Error("id conversion err:%v", err)
		return
	}

	conn, err := upGrader.Upgrade(w, r, w.Header())
	if err != nil {
		logs.Error("upgrade err:%v", err)
		return
	}
	client := &Client{conn: conn, HandPokers: make([]int, 0, 21), UserInfo: &UserInfo{}}
	if userId != 0 && userName != "" {
		client.UserInfo.UserId = UserId(userId)
		client.UserInfo.Username = userName

		client = AddClient(client)
		//client.setCloseHandler()
		go client.readPump()
		return
	}
	logs.Error("user need login first")
	client.conn.Close()
}
