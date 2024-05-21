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
	SetFrom  url.Values
}

func copyValues(Form url.Values) url.Values {
	to := make(url.Values)
	for k, v := range Form {
		to[k] = append([]string{}, v...)
	}
	return to
}

func (m *Parameter) CopyParameter() *Parameter {
	parameter := &Parameter{Path: m.Path, Form: copyValues(m.Form), PostForm: copyValues(m.PostForm), SetFrom: copyValues(m.SetFrom)}
	return parameter
}

func NewParameter(re *http.Request) *Parameter {
	re.ParseForm()
	path := re.URL.Path
	return &Parameter{Path: path, Form: re.Form, PostForm: re.PostForm, SetFrom: make(url.Values)}
}
func (m *Parameter) GetString(key string) string {
	if m.SetFrom != nil {
		if m.SetFrom.Has(key) {
			return m.SetFrom.Get(key)
		}
	}
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

func (m *Parameter) GetVString(keys ...string) string {
	for _, key := range keys {
		v := m.GetString(key)
		if len(v) > 0 {
			return v
		}
	}
	return ""
}

func (m *Parameter) SetString(key string, value string) {
	m.SetFrom[key] = []string{value}
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
func (m *Parameter) SetIndex(index int) {
	m.SetString("index", strconv.Itoa(index))
}
func (m *Parameter) GetIndex() int {
	return m.GetInt("index")
}

func GetUsername(re *Parameter) string {
	username := re.GetVString("id", "username")
	return username
}

type RegisterHandle func(parameter *Parameter) any
