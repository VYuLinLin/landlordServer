package service

var (
	ClientList = make([]*Client, 0, 100)
)

func AddClient(c *Client) *Client {
	for i := 0; i < len(ClientList); i++ {
		if ClientList[i].UserInfo.UserId == c.UserInfo.UserId {
			ClientList[i].conn, c = c.conn, ClientList[i]
			return c
		}
	}
	ClientList = append(ClientList, c)
	return c
}
func GetClient(id UserId) *Client  {
	for i := 0; i < len(ClientList); i++ {
		if ClientList[i].UserInfo.UserId == id {
			return ClientList[i]
		}
	}
	return nil
}
func DeleteClient(id UserId)  {
	for i := 0; i < len(ClientList); i++ {
		if ClientList[i].UserInfo.UserId == id {
			ClientList = append(ClientList[0:i], ClientList[i+1:]...)
			return
		}
	}
}