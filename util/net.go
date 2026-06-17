package util

import (
	"net/http"
	"strconv"
	"strings"
)

func GetUsername(re *http.Request) string {
	username := re.FormValue("id")
	if len(username) == 0 {
		username = re.FormValue("username")
	}
	return username
}

func GetGroupId(re *http.Request) string {
	group := re.FormValue("groupId")
	if len(group) == 0 {
		group = re.FormValue("GroupId")
	}
	return group
}

func GetGroupIds(re *http.Request) []string {
	group := re.FormValue("groupId")
	if len(group) == 0 {
		group = re.FormValue("GroupId")
	}
	if len(group) > 0 {
		v := strings.TrimSpace(strings.Trim(group, ","))
		return strings.Split(v, ",")
	}
	return []string{}
}

func GetLiveTime(re *http.Request) int {
	liveTime := re.FormValue("liveTime")
	if len(liveTime) == 0 {
		return 0
	}
	liveTimeIni, err := strconv.Atoi(liveTime)
	if err == nil {
		return liveTimeIni
	}
	return 0
}

func GetMessage(re *http.Request) string {
	msg := re.FormValue("msg")
	if len(msg) == 0 {
		msg = re.FormValue("message")
	}
	return msg
}

func HttpCross(w http.ResponseWriter) {
	h := w.Header()
	h.Add("Access-Control-Allow-Origin", "*")
	h.Add("Content-Type", "text/html; charset=utf-8")
}
