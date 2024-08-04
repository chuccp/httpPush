package cluster0

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
)

type User struct {
	MachineId string
	UserId    string
}

func NewUser(MachineId string, UserId string) *User {
	return &User{MachineId: MachineId, UserId: UserId}
}

type Response struct {
	Code int
	Msg  string
}

func successResponse() *Response {
	return &Response{Code: 200, Msg: "success"}
}
func failResponse(msg string) *Response {
	return &Response{Code: 500, Msg: msg}
}

type clusterSendMessage struct {
	ous         []user.IOrderUser
	index       int
	msg         message.IMessage
	exMachineId []string
	machineId   string
}
