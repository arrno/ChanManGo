package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	fmt.Println("[SUBSCRIBER]")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:8080/subscribe", nil)
	if err != nil {
		panic(err)
	}
	defer c.CloseNow()

	for {
		fmt.Println("[CLIENT] subscribing for signal")
		var data map[string]any
		err = wsjson.Read(ctx, c, &data)
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
}
