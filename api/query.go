package api

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"log"
	"net/http"
)

type Query struct {
	context *core.Context
	server  core.IHttpServer
}

func (query *Query) query(w http.ResponseWriter, re *http.Request) {
	path := re.URL.Path
	parameter := core.NewParameter(path, re)
	value := query.context.Query(parameter)
	data, _ := json.Marshal(value)
	w.Write(data)
}

func (query *Query) Init() {
	query.AddQuery("/queryUser", query.queryUser)
}
func (query *Query) AddQuery(handleName string, handle core.RegisterHandle) {
	query.context.RegisterHandle(handleName, handle)
	query.server.AddHttpRoute(handleName, query.query)
}

func (query *Query) queryUser(parameter *core.Parameter) any {
	log.Println("queryUser")
	var u User
	machineInfoId, ok := query.context.GetHandle("machineInfoId")
	if ok {
		u.Machine = machineInfoId(parameter)
	}
	u.Conn = make([]*Conn, 0)
	data, _ := json.Marshal(parameter)
	log.Println(string(data))
	id := core.GetUsername(parameter)
	log.Println("id", id)
	if len(id) > 0 {
		us, fa := query.context.GetUser(id)
		if fa {
			for _, iUser := range us {
				u.Username = iUser.GetUsername()
				u.Conn = append(u.Conn, newConn(iUser.GetRemoteAddress(), iUser.LastLiveTime().Format(util.TimestampFormat), iUser.CreateTime().Format(util.TimestampFormat)))
			}
		}
	}
	return &u
}
func NewQuery(context *core.Context, server core.IHttpServer) *Query {
	query := &Query{context: context, server: server}
	return query
}

type User struct {
	Username string
	Conn     []*Conn
	Machine  interface{}
}
type Conn struct {
	RemoteAddress string
	LastLiveTime  string
	CreateTime    string
}

func newConn(RemoteAddress string, LastLiveTime string, CreateTime string) *Conn {
	return &Conn{RemoteAddress, LastLiveTime, CreateTime}
}
