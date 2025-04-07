package websockets

import (
	"MyBloge/model"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Pool   *Pool
	RoomID string // 当前所属的聊天室 ID
	Mu     sync.Mutex
}

// ReadMessageFromRoom 从群聊聊天室中获得值
func (c *Client) ReadMessageFromRoom(roomID string, userId string) {
	defer func() {
		// 这里不能加锁，否则会形成死锁
		c.Pool.Unregister <- c
		err := c.Conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()
	for {
		messageType, messageInfo, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break //*** break 是退出这个循环， return 是退出这个函数， 使用break为了正常关闭连接
		}
		messageId := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(time.Now().String())).String()
		message := model.Message{
			MessageType: messageType,
			Message:     string(messageInfo),
			MessageId:   messageId,
			SenderId:    userId,
			RoomId:      c.RoomID,
			Timestamp:   time.Now().Unix(),
		}
		// 一条来发送，一条来存储  ***
		c.Pool.Broadcast <- message
		c.Pool.Broadcast <- message
		fmt.Println("Message sent", message.Message)
	}
}
