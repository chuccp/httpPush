package cluster

type User struct {
	MachineId string
	UserId    string
}

func NewUser(MachineId string, UserId string) *User {
	return &User{MachineId: MachineId, UserId: UserId}
}

type Response struct {
	code int
	msg  string
}

func successResponse() *Response {
	return &Response{code: 200, msg: "success"}
}
func failResponse(msg string) *Response {
	return &Response{code: 500, msg: msg}
}
