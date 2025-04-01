package controllers

import (
	"MyBloge/db"
	"MyBloge/global"
	"MyBloge/model"
	"MyBloge/utils"
	"MyBloge/websockets"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

// CreateChatRoom 根据用户Id就可以创建聊天室了，这个就是一个群聊
func CreateChatRoom(context *gin.Context) {
	var request utils.CreateRoomRequest
	if err := context.ShouldBind(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room, err := db.CreateChatRoomDB(request.RoomName, request.CreatorId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"chat_room": room})
}

func GetChatRooms(context *gin.Context) {
	rooms, err := db.GetAllChatRoom()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"chat_rooms": rooms})
}

func GetChatRoom(context *gin.Context) {
	roomId := context.Param("roomId")
	var room model.ChatRoom
	room, err := db.GetRoomById(roomId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"chat_room": room})
}

// JoinChatRoom 加入一个群聊，开始监听消息，数据库里记录
func JoinChatRoom(context *gin.Context) {
	userId := context.Query("userId")
	roomId := context.Query("roomId")
	// 将用户加入群聊
	err := db.AddUserToChatRoom(userId, roomId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 升级为WebSocket连接
	conn, err := websockets.Upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 如果全局连接池为空，初始化它
	if global.GlobalPool == nil {
		global.GlobalPool = websockets.NewPool()
		go global.GlobalPool.Start()
		go global.GlobalPool.HeartBeatCheck()
	}
	// 创建一个新的客户端实例，避免使用全局变量
	client := &websockets.Client{
		ID:     userId,
		Conn:   conn,
		Pool:   global.GlobalPool,
		RoomID: roomId,
		Mu:     sync.Mutex{},
	}
	// 启动goroutine持续监听消息
	go func(c *websockets.Client) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("goroutine panicked: %v", r)
			}
		}()
		for {
			select {
			case message := <-c.Pool.Broadcast:
				log.Println("收到消息：", message.Message)
				if err := db.AddMessageToDb(message); err != nil {
					log.Printf("数据库操作失败: %v", err)
					// 不退出goroutine，继续监听
					continue
				}
			}
		}
	}(client)
	global.GlobalPool.Register <- client
	client.ReadMessageFromRoom(roomId, userId)
	context.JSON(http.StatusCreated, gin.H{"success": "Join successfully! Welcome!"})
}

// LeaveOutCharRoom 停止对这个room的监听, 从数据库中去除这个user
func LeaveOutCharRoom(context *gin.Context) {
	roomId := context.Query("roomId")
	userId := context.Query("userId")
	err := db.LeaveOutChatRoom(roomId, userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	// 注销client 在pool中会实现,前端执行这个之前要先执行退出websocket连接
	context.JSON(http.StatusCreated, gin.H{"success": "Left successfully!"})
}

// DeleteChatRoom 直接删除一个聊天室, 只有创建者有权利
func DeleteChatRoom(context *gin.Context) {
	roomID := context.Query("roomId")
	// 清除user ...
	if err := db.DeleteChatRoomDb(roomID); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"success": "Delete successfully!"})
}

// SendMessage 发送一个消息在聊天室中，数据库会记录这条记录，发送消息实际上在前端运行websocket，后端只要记录这条消息即可
func SendMessage(context *gin.Context) {
	var request utils.SendMessageRequest
	if err := context.ShouldBind(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	message := model.Message{
		MessageId:   request.MessageId,
		RoomId:      request.RoomId,
		MessageType: request.MessageType,
		SenderId:    request.SenderId,
		Message:     request.MessageContent,
		Timestamp:   request.Timestamp,
	}
	if err := db.AddMessageToDb(message); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"success": "Add Message successfully!"})
}

// GetMessageHistory 拿出所有的聊天记录，当然可以根据你加入的时间获得相应的聊天记录
func GetMessageHistory(context *gin.Context) {
	roomID := context.Param("roomId")
	Messages, err := db.GetRoomMessage(roomID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": Messages})
}

func CreatePrivateRoom(context *gin.Context) {
	var request utils.CreatePrivateRoomRequest
	if err := context.ShouldBind(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room, err := db.CreatePrivateChatRoomDB(request)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"success": room})
}

// StartPrivateChat 没什么区别
func StartPrivateChat(context *gin.Context) {
	userAId := context.Query("userAId") //代表本人
	userBId := context.Query("userBId") //代表好友
	roomId := context.Query("roomId")
	//先检查这2个是不是已经加入过了，不能重复
	allIn, err := db.FindUserInRoom(userAId, userBId, roomId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if !allIn {
		// 将2个用户都加入群聊
		err = db.AddUserToChatRoom(userAId, roomId)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = db.AddUserToChatRoom(userBId, roomId)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	// 升级为WebSocket连接
	conn, err := websockets.Upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 如果全局连接池为空，初始化它
	if global.GlobalPool == nil {
		global.GlobalPool = websockets.NewPool()
		go global.GlobalPool.Start()
		go global.GlobalPool.HeartBeatCheck()
	}
	// 创建一个新的客户端实例，避免使用全局变量
	client := &websockets.Client{
		ID:     userAId,
		Conn:   conn,
		Pool:   global.GlobalPool,
		RoomID: roomId,
		Mu:     sync.Mutex{},
	}
	// 启动goroutine持续监听消息
	go func(c *websockets.Client) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("goroutine panicked: %v", r)
			}
		}()
		for {
			select {
			case message := <-c.Pool.Broadcast:
				log.Println("收到消息：", message.Message)
				if err := db.AddMessageToDb(message); err != nil {
					log.Printf("数据库操作失败: %v", err)
					// 不退出goroutine，继续监听
					continue
				}
			}
		}
	}(client)
	global.GlobalPool.Register <- client
	client.ReadMessageFromRoom(roomId, userAId)
	context.JSON(http.StatusCreated, gin.H{"success": "Join successfully! Welcome!"})
}

// GetPrivateChats 获得所有的私人聊天室
func GetPrivateChats(context *gin.Context) {
	privateRooms, err := db.GetAllPrivateRoom()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": privateRooms})
}

// GetPrivateChatHistory 从私人聊天室中获得消息记录
func GetPrivateChatHistory(context *gin.Context) {
	roomId := context.Param("roomId")
	messages, err := db.GetRoomMessage(roomId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": messages})
}
