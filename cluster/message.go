package cluster

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
