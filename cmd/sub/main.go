package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	var wg sync.WaitGroup
	for i := range 3 {
		wg.Add(1)
		makeSubscriber(&wg, fmt.Sprintf("%d", i))
	}
	wg.Wait()
}

func makeSubscriber(wg *sync.WaitGroup, id string) {
	go func() {
		defer wg.Done()
		fmt.Println("[SUBSCRIBER] -> " + id)

		headers := http.Header{}
		headers.Set("Authorization", "Bearer auth-token")
		headers.Set("channel", "MyChannel")
		if id == "1" {
			headers.Set("topics", "topicOne")
		} else {
			headers.Set("topics", "topicTwo")
		}
		c, _, err := websocket.Dial(context.Background(), "ws://localhost:8080/subscribe", &websocket.DialOptions{HTTPHeader: headers})
		if err != nil {
			panic(err)
		}
		defer c.CloseNow()

		for {
			fmt.Println("[CLIENT] subscribing for signal: " + id)
			var data map[string]any
			// ctx, _ := context.WithTimeout(context.Background(), time.Minute)
			err = wsjson.Read(context.Background(), c, &data)
			if err != nil {
				panic(err)
			}
			b, err := json.MarshalIndent(data, "", "    ")
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			fmt.Println(string(b))
		}
		c.Close(websocket.StatusNormalClosure, "")
	}()
}
