package core

import (
	"net/http"
	"net/url"
	"strconv"
)

type Parameter struct {
	Path     string
	Form     url.Values
	PostForm url.Values
}

func NewParameter(Path string, re *http.Request) *Parameter {
	re.ParseForm()
	return &Parameter{Path: Path, Form: re.Form, PostForm: re.PostForm}
}
func (m *Parameter) GetString(key string) string {
	if m.Form != nil {
		if m.Form.Has(key) {
			return m.Form.Get(key)
		}
	}
	if m.PostForm != nil {
		if m.PostForm.Has(key) {
			return m.PostForm.Get(key)
		}
	}
	return ""
}
func (m *Parameter) GetInt(key string) int {
	v := m.GetString(key)
	if len(v) > 0 {
		num, err := strconv.Atoi(v)
		if err != nil {
			return 0
		} else {
			return num
		}
	}
	return 0
}

func GetUsername(re *Parameter) string {
	username := re.GetString("id")
	if len(username) > 0 {
		return username
	}
	username = re.GetString("username")
	return username
}

type RegisterHandle func(parameter *Parameter) any
