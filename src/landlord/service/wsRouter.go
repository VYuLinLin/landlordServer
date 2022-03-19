package service

const (
	UserUpdate    = "user/update"     // 更新用户信息
	RoomList      = "room/list"       // 房间列表
	RoomJoinSelf  = "room/join/self"  // 进入房间
	RoomJoinOther = "room/join/other" // 其他玩家进入房间
	RoomLeave     = "room/leave"      // 离开房间
	TableInfo     = "table/info"      // 桌子信息
	TableStatus   = "table/status"    // 桌子状态

	TableUpdate      = "table/data/update" // 桌子状态
	TableJoin        = "table/join"        // 进入桌子
	PlayerReady      = "player/ready"      // 玩家准备
	PlayerDeal       = "player/deal"        // 发牌
	TableCallPoints  = "table/callPoints"  // 抢地主
	PlayerCallPoints = "player/callPoints" // 玩家叫分
	TableHoleCards = "table/holeCards" // 显示底牌
)
