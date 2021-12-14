package service

var (
	ClientList = make([]*Client, 0, 100)
)

func AddClient(c *Client)  {
	for i := 0; i < len(ClientList); i++ {
		if ClientList[i].UserInfo.UserId == c.UserInfo.UserId {
			ClientList[i] = c
			return
		}
	}
	ClientList = append(ClientList, c)
}

func DeleteClient(id UserId)  {
	for i := 0; i < len(ClientList); i++ {
		if ClientList[i].UserInfo.UserId == id {
			ClientList = append(ClientList[0:i], ClientList[i+1:]...)
			return
		}
	}
}