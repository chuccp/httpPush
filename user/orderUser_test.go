package user

import (
	"github.com/chuccp/httpPush/message"
	"sort"
	"testing"
	"time"
)

type TestOrderUser struct {
	id         int
	priority   int
	machineID  string
	orderTime  time.Time
	orderValue int
}

func (u *TestOrderUser) GetPriority() int {
	return u.priority
}
func (u *TestOrderUser) GetMachineId() string {
	return u.machineID
}
func (u *TestOrderUser) GetOrderTime() *time.Time {
	return &u.orderTime
}
func (u *TestOrderUser) WriteMessage(iMessage message.IMessage, writeFunc WriteCallBackFunc) {
	// 实现省略，根据实际需求编写
}

func TestByAsc_Less(t *testing.T) {
	users := []IOrderUser{
		&TestOrderUser{id: 1, priority: 3, orderTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		&TestOrderUser{id: 2, priority: 2, orderTime: time.Date(2022, 1, 1, 0, 0, 1, 0, time.UTC)},
		&TestOrderUser{id: 3, priority: 2, orderTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		&TestOrderUser{id: 4, priority: 1, orderTime: time.Date(2023, 1, 1, 0, 0, 2, 0, time.UTC)},
	}
	sort.Sort(ByAsc(users))

	for _, v := range users {
		t.Log(v)
	}

}
