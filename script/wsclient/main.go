package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8086", Path: "/ws", RawQuery: "id=user_a"}
	fmt.Println("connect:", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()
	fmt.Println("connected")

	// Clear welcome
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	conn.ReadMessage()
	conn.SetReadDeadline(time.Time{})

	// HTTP send
	go func() {
		time.Sleep(500 * time.Millisecond)
		resp, err := http.Get("http://127.0.0.1:8086/sendmsg?username=user_a&msg=hello-from-go-client")
		if err != nil {
			log.Fatal("http:", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println("http response:", string(body))
	}()

	// Wait for WS message
	fmt.Println("waiting for message...")
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("ws read:", err)
	}
	fmt.Printf("received: %s\n", string(msg))
}
