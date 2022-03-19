package service

import (
	"fmt"
	"landlord/common"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

type TableId int

const (
	GameWaitting       = iota
	GamePushCard       // 发牌
	GameCallScore      // 叫分
	GameShowBottomCard // 显示底牌
	GamePlaying        // 出牌
	GameEnd            // 结束
)

type Table struct {
	Lock         sync.RWMutex
	TableId      TableId
	State        int
	Creator      *Client
	TableClients map[UserId]*Client
	GameManage   *GameManage
}
type TableData struct {
	RoomId  RoomId        `json:"room_id"`
	TableId TableId       `json:"table_id"`
	State   int           `json:"state"`
	Creator UserId        `json:"creator"`
	Clients []interface{} `json:"clients"`
}

// 获取当前用户所在桌子的信息
func (t *Table) getTableData(c *Client) {
	var msg = "ok"
	if msg != "ok" {
		c.sendMsg(TableInfo, 500, msg)
	} else {
		var data = &TableData{
			RoomId:  c.Room.RoomId,
			TableId: t.TableId,
			State:   t.State,
			Creator: t.Creator.UserInfo.UserId,
		}
		//for _, user := range t.TableClients {
		//	data.Clients = nil
			var nextId UserId
			for _, client := range t.TableClients {
				if client.Next != nil {
					nextId = client.Next.UserInfo.UserId
				}
				avatarUrl := ""
				userData := map[string]interface{}{
					"user_name":  client.UserInfo.Username,
					"user_id":    client.UserInfo.UserId,
					"coin":       client.UserInfo.Coin,
					"role":       client.UserInfo.Role,
					"ready":     client.Ready,
					"status":     client.Status,
					"next_id":    nextId,
					"avatar_url": avatarUrl,
				}
				data.Clients = append(data.Clients, userData)
			}

			c.sendMsg(TableInfo, 200, data)
		//}
	}
}

//客户端加入牌桌
func (t *Table) joinTable(c *Client) {
	var msg = "ok"
	defer func() {
		if msg != "ok" {
			c.sendMsg(RoomJoinSelf, 500, msg)
		} else {
			// 通知自己
			data := map[string]int{
				"room_id":  int(c.Room.RoomId),
				"table_id": int(t.TableId),
			}
			c.sendMsg(RoomJoinSelf, 200, data)
			c.Status = UNREADY

			// 通知其他玩家
			var nextId UserId
			if c.Next != nil {
				nextId = c.Next.UserInfo.UserId
			}
			userData := map[string]interface{}{
				"user_name":  c.UserInfo.Username,
				"user_id":    c.UserInfo.UserId,
				"coin":       c.UserInfo.Coin,
				"role":       c.UserInfo.Role,
				"status":     c.Status,
				"next_id":		nextId,
				"avatar_url": "",
			}
			for _, client := range t.TableClients {
				if client.UserInfo.UserId == c.UserInfo.UserId {
					continue
				}
				client.sendMsg(RoomJoinOther, 200, userData)
			}
		}
	}()
	//t.Lock.Lock()
	//defer t.Lock.Unlock()
	if _, ok := t.TableClients[c.UserInfo.UserId]; ok {
		return
	}
	if len(t.TableClients) > 3 {
		msg = fmt.Sprintf("Player[%d] JOIN Table[%d] FULL", c.UserInfo.UserId, t.TableId)

		logs.Error(msg)
		return
	}
	logs.Debug("[%v] user [%v] request join t", c.UserInfo.UserId, c.UserInfo.Username)

	c.Table = t
	//c.Ready = true
	for _, client := range t.TableClients {
		if client.Next == nil {
			client.Next = c
			break
		}
	}
	t.TableClients[c.UserInfo.UserId] = c
	if len(t.TableClients) == 1 {
		t.Creator = c
	}
	//t.syncUser()
	//if len(t.TableClients) == 3 {
	//	c.Next = t.Creator
	//	t.State = GameCallScore
	//	t.dealPoker()
	//} else
	if c.Room.AllowRobot {
		//go t.addRobot(c.Room)
		logs.Debug("robot join ok")
	}
	if len(t.TableClients) == 3 {
		c.Next = t.Creator
		t.State = GameWaitting
		t.GameManage.FirstCallScore = t.Creator
	}
}

func (t *Table) leaveTable(c *Client) {
	if c.Status == INVALID || !c.Ready {
		t.State = GameWaitting
		delete(t.TableClients, c.UserInfo.UserId)
		c.Table = nil
		c.Next = nil
		if len(t.TableClients) == 0 {
			t.Creator = nil
			return
		}
		for _, client := range t.TableClients {
			if len(t.TableClients) == 1 {
				client.Next = nil
				t.Creator = client
			}else if client.Next == c {
				client.Next = nil
			} else if t.Creator == c {
				t.Creator = client
			}
		}
	}
}

//加入机器人
func (t *Table) addRobot(room *Room) {
	logs.Debug("robot [%v] join t", fmt.Sprintf("ROBOT-%d", len(t.TableClients)))
	if len(t.TableClients) < 3 {
		client := &Client{
			Room:       room,
			HandPokers: make([]int, 0, 21),
			UserInfo: &UserInfo{
				UserId:   t.getRobotID(),
				Username: fmt.Sprintf("ROBOT-%d", len(t.TableClients)),
				Coin:     10000,
			},
			IsRobot:  true,
			toRobot:  make(chan []interface{}, 3),
			toServer: make(chan []interface{}, 3),
		}
		go client.runRobot()
		t.joinTable(client)
	}
}

//生成随机robotID
func (t *Table) getRobotID() (robot UserId) {
	time.Sleep(time.Microsecond * 10)
	rand.Seed(time.Now().UnixNano())
	robot = UserId(rand.Intn(10000))
	t.Lock.RLock()
	defer t.Lock.RUnlock()
	if _, ok := t.TableClients[robot]; ok {
		return t.getRobotID()
	}
	return
}

/**
开始游戏
游戏顺序: 等待 =》 准备 =》 发牌 =》 抢地主（叫分） =》 显示底牌 =》 出牌 =》 游戏结束 =》 等待
*/
func (t *Table) gameStart() {
	if len(t.TableClients) < 3 {
		return
	}
	for _, client := range t.TableClients {
		if !client.Ready {
			return
		}
	}
	t.dealPoker()
	time.Sleep(3 * 1e9)
	t.callPointsStart()
}

// 发牌
func (t *Table) dealPoker() {
	logs.Debug("deal poker")
	t.GameManage.Pokers = make([]int, 0)
	for i := 0; i < 54; i++ {
		t.GameManage.Pokers = append(t.GameManage.Pokers, i)
	}
	t.ShufflePokers()
	for i := 0; i < 17; i++ {
		for _, client := range t.TableClients {
			client.HandPokers = append(client.HandPokers, t.GameManage.Pokers[len(t.GameManage.Pokers)-1])
			t.GameManage.Pokers = t.GameManage.Pokers[:len(t.GameManage.Pokers)-1]
		}
	}
	//response := make([]interface{}, 0, 3)
	//response = append(append(append(response, common.ResDealPoker), t.GameManage.FirstCallScore.UserInfo.UserId), nil)
	for _, client := range t.TableClients {
		sort.Ints(client.HandPokers)
		//response[len(response)-1] = client.HandPokers
		client.sendMsg(PlayerDeal, 200, client.HandPokers)
	}
}

// ShufflePokers 洗牌
func (t *Table) ShufflePokers() {
	logs.Debug("ShufflePokers")
	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := len(t.GameManage.Pokers)
	for i > 0 {
		randIndex := r.Intn(i)
		t.GameManage.Pokers[i-1], t.GameManage.Pokers[randIndex] = t.GameManage.Pokers[randIndex], t.GameManage.Pokers[i-1]
		i--
	}
}

// 开始抢地主（叫分）
func (t *Table) callPointsStart() {
	t.State = GameCallScore
	//r := rand.Intn(3) // 生成3以内的数字
	//ids := make([]UserId, 0, 3)
	//for id := range t.TableClients {
	//	ids = append(ids, id)
	//}
	player := t.GameManage.FirstCallScore
	//userId := t.GameManage.FirstCallScore.UserInfo.UserId
	for _, client := range t.TableClients {
		client.sendMsg(TableCallPoints, 200, player.UserInfo.UserId)
	}
	//t.GameManage.Turn = t.GameManage.FirstCallScore.Next
	//go func() {
	//	for i := 0; i < 3; i++ {
			time.Sleep(10e9)
			player = t.GameManage.FirstCallScore
			if t.State == GamePlaying || player.IsCalled {
				return
			}
			t.callPoints(player.UserInfo.UserId, 0)
	//	}
	//}()
}

// 抢地主（叫分）
func (t *Table) callPoints(id UserId, score int) {
	data := map[string]int{
		"user_id": int(id),
		"score":   score,
	}
	var playerClient *Client
	for _, client := range t.TableClients {
		if client.UserInfo.UserId == id {
			client.IsCalled = true
			playerClient = client
			break
		}
		client.sendMsg(PlayerCallPoints, 200, data)
	}

	//t.GameManage.MaxCallScoreTurn.IsCalled = true
	//t.GameManage.MaxCallScoreTurn = playerClient.Next
	if score > t.GameManage.MaxCallScore {
		t.GameManage.MaxCallScore = score
		t.GameManage.MaxCallScoreTurn = playerClient
	}

	isCallEnd := score == 3 || t.allCalled()
	if isCallEnd {
		logs.Debug("call score end")
		t.callEnd()
		return
	}
	// 通知下一位叫地主
	nextUser := playerClient.Next
	for _, client := range t.TableClients {
		client.sendMsg(TableCallPoints, 200, nextUser.UserInfo.UserId)
	}
	time.Sleep(10e9)
	if !nextUser.IsCalled {
		t.callPoints(nextUser.UserInfo.UserId, 0)
	}
}

// 叫分阶段结束
func (t *Table) callEnd() {
	t.State = GamePlaying
	if t.GameManage.MaxCallScoreTurn == nil || t.GameManage.MaxCallScore == 0 {
		t.GameManage.MaxCallScoreTurn = t.GameManage.FirstCallScore
		t.GameManage.MaxCallScore = 1
		//return
	}
	t.GameManage.FirstCallScore = t.GameManage.FirstCallScore.Next
	landLord := t.GameManage.MaxCallScoreTurn
	landLord.UserInfo.Role = RoleLandlord
	//t.GameManage.Turn = landLord
	for _, poker := range t.GameManage.Pokers {
		landLord.HandPokers = append(landLord.HandPokers, poker)
	}
	//res := []interface{}{common.ResShowPoker, landLord.UserInfo.UserId, t.GameManage.Pokers}
	data := map[string]interface{}{
		"table_state": t.State,
		"landLord_id":  landLord.UserInfo.UserId,
		"pokers": t.GameManage.Pokers,
	}
	for _, c := range t.TableClients {
		c.sendMsg(TableHoleCards, 200, data)
	}
}

func (t *Table) stateUpdate(state int) {
	if t.State != state {
		t.State = state
		for _, client := range t.TableClients {
			client.sendMsg(TableStatus, 200, state)
		}
	}
}

func (t *Table) allCalled() bool {
	for _, client := range t.TableClients {
		if !client.IsCalled {
			return false
		}
	}
	return true
}

//一局结束
func (t *Table) gameOver(client *Client) {
	coin := t.Creator.Room.EntranceFee * t.GameManage.MaxCallScore * t.GameManage.Multiple
	t.State = GameEnd
	for _, c := range t.TableClients {
		res := []interface{}{common.ResGameOver, client.UserInfo.UserId}
		if client == c {
			res = append(res, coin*2-100)
		} else {
			res = append(res, coin)
		}
		for _, cc := range t.TableClients {
			if cc != c {
				userPokers := make([]int, 0, len(cc.HandPokers)+1)
				userPokers = append(append(userPokers, int(cc.UserInfo.UserId)), cc.HandPokers...)
				res = append(res, userPokers)
			}
		}
		//c.sendMsg(res)
	}
	logs.Debug("t[%d] game over", t.TableId)
}

func (t *Table) chat(client *Client, msg string) {
	//res := []interface{}{common.ResChat, client.UserInfo.UserId, msg}
	//for _, c := range t.TableClients {
	//	c.sendMsg(res)
	//}
}

func (t *Table) reset() {
	t.GameManage = &GameManage{
		FirstCallScore:   t.GameManage.FirstCallScore,
		Turn:             nil,
		MaxCallScore:     0,
		MaxCallScoreTurn: nil,
		LastShotClient:   nil,
		Pokers:           t.GameManage.Pokers[:0],
		LastShotPoker:    t.GameManage.LastShotPoker[:0],
		Multiple:         1,
	}
	t.State = GameWaitting
	//if t.Creator != nil {
	//	t.Creator.sendMsg([]interface{}{common.ResRestart})
	//}
	for _, c := range t.TableClients {
		c.reset()
	}
	if len(t.TableClients) == 3 {
		t.dealPoker()
	}
}

//同步用户信息
func (t *Table) syncUser() {
	logs.Debug("sync user")
	response := make([]interface{}, 0, 3)
	response = append(append(response, common.ResJoinTable), t.TableId)
	tableUsers := make([][2]interface{}, 0, 2)
	current := t.Creator
	for i := 0; i < len(t.TableClients); i++ {
		tableUsers = append(tableUsers, [2]interface{}{current.UserInfo.UserId, current.UserInfo.Username})
		current = current.Next
	}
	response = append(response, tableUsers)
	for _, client := range t.TableClients {
		client.sendMsg(UserUpdate, 200, response)
	}
}
