package service

type GameManage struct {
	Turn             *Client
	FirstCallScore   *Client //每局轮转
	MaxCallScore     int     //最大叫分
	MaxCallScoreTurn *Client
	LastShotClient   *Client
	Pokers           []int
	LastShotPoker    []int
	Multiple         int //加倍
}
