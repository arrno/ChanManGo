package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	fmt.Println("[PUBLISHER]")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer auth-token")
	headers.Set("channel", "MyChannel")
	c, _, err := websocket.Dial(ctx, "ws://localhost:8080/publish", &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		panic(err)
	}
	defer c.CloseNow()

	for i := range 10 {
		fmt.Printf("[PUBLISHER] sending signal %d\n", i)
		data := map[string]any{
			"Hello": "Chanban",
			"Val":   i,
		}
		if i%2 == 0 {
			err = writeMessage(c, "topicOne", data)
		} else {
			err = writeMessage(c, "topicTwo", data)
		}
		if err != nil {
			// TODO should inform caller success or failure
			panic(err)
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
	c.Close(websocket.StatusNormalClosure, "")
}

func writeMessage(c *websocket.Conn, topic string, msg map[string]any) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	wrapped := map[string]any{
		"topic":   topic,
		"message": bytes,
	}
	return wsjson.Write(context.Background(), c, &wrapped)
}
