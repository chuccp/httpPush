package cluster

import "github.com/chuccp/httpPush/util"

type Client struct {
	machine *Machine
	request *util.Request
}

func NewClient() *Client {
	return &Client{request: util.NewRequest()}
}
func (client *Client) initial() {

}
