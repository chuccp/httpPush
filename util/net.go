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
func GetGroupIds(re *http.Request) []string {
	group := re.FormValue("groupId")
	if len(group) == 0 {
		group = re.FormValue("GroupId")
	}
	if len(group) == 0 {
		return strings.Split(group, ",")
	}
	return []string{}
}
func GetGroupId(re *http.Request) string {
	group := re.FormValue("groupId")
	if len(group) == 0 {
		group = re.FormValue("GroupId")
	}

	return group
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

func GetStart(re *http.Request) int {
	value := re.FormValue("start")
	start, err := strconv.Atoi(value)
	if err == nil {
		return start
	}
	return 0
}
func GetSize(re *http.Request) int {
	value := re.FormValue("size")
	start, err := strconv.Atoi(value)
	if err == nil {
		return start
	}
	return 10
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
func HttpCrossChunked(w http.ResponseWriter) {
	h := w.Header()
	HttpCross(w)
	h.Add("Transfer-Encoding", "chunked")
}
