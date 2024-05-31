package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestName0000(t *testing.T) {
	// 创建一个HTTP客户端
	client := &http.Client{}

	// 创建一个GET请求
	req, err := http.NewRequest("GET", "http://127.0.0.1:5656", nil)
	if err != nil {
		fmt.Println("创建请求时出错:", reflect.TypeOf(err))
		return
	}

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {

		//fmt.Println("创建请求时出错:", reflect.TypeOf(rr.Err))
		fmt.Println("发送请求时出错:", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应的内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应时出错:", err)
		return
	}

	// 打印响应的内容
	fmt.Println(string(body))
}
func TestName03232(t *testing.T) {
	//bytes.Buffer{}

}
