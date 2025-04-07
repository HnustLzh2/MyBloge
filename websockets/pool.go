package websockets

import (
	"MyBloge/model"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool            //所以的Clients
	Rooms      map[string]map[*Client]bool // 每个聊天室ID对应一个客户端集合*** important
	Broadcast  chan model.Message
	Mu         sync.Mutex // 保护对 Clients 的并发访问
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Rooms:      make(map[string]map[*Client]bool),
		Broadcast:  make(chan model.Message),
	}
}

// HeartBeat 用于心跳包检测
type HeartBeat struct {
	message     string
	timestamp   int64 //高精度时间戳，通常在grpc使用
	sessionId   string
	nonce       int
	cpuUsage    float64
	memoryUsage float64
	loadState   string
}

const (
	heartBeatTimeDuration = 10 * time.Second
)

func (pool *Pool) HeartBeatCheck() {
	ticker := time.NewTicker(heartBeatTimeDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			pool.Mu.Lock()
			// 获取CPU使用率
			percent, err := cpu.Percent(time.Second, false)
			if err != nil {
				log.Printf("Error getting CPU percent: %v", err)
				pool.Mu.Unlock()
				return
			}
			// 获取内存信息
			memInfo, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("Error getting memory info: %v", err)
				pool.Mu.Unlock()
				return
			}
			// 计算负载状态
			var loadState string
			loadState = "low"
			if memInfo.UsedPercent > 0.5 && memInfo.UsedPercent < 0.8 {
				loadState = "middle"
			} else if memInfo.UsedPercent >= 0.8 {
				loadState = "busy"
			}
			// 创建心跳包
			heartBeat := HeartBeat{
				message:     "heartbeat",
				timestamp:   time.Now().Unix(),
				sessionId:   "",
				nonce:       1000 + rand.Intn(9000), // 1000 到 10000 (四位数)
				cpuUsage:    percent[0],             // CPU占用
				memoryUsage: memInfo.UsedPercent,    // 内存占用
				loadState:   loadState,
			}
			// 向所有客户端发送心跳包,如果发送心跳包的时候断开了链接，就会出现错误
			for client := range pool.Clients {
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte(heartBeat.message))
				if err != nil {
					log.Printf("Error sending heartbeat to client: %v", err)
					if err := client.Conn.Close(); err != nil {
						fmt.Printf("Error closing client: %v", err)
						break //*** break 是退出这个循环， return 是退出这个函数，使用break为了正常关闭连接
					}
					break
				}
				err = client.Conn.WriteJSON(&heartBeat)
				if err != nil {
					log.Printf("Error sending heartbeat to client: %v", err)
					if err := client.Conn.Close(); err != nil {
						fmt.Printf("Error closing client: %v", err)
						break
					}
					break
				}
			}
			pool.Mu.Unlock()
		}
	}
}

// Start 监听加入房间，离开房间，监听消息的传递
func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Mu.Lock()
			fmt.Println(client)
			pool.Clients[client] = true
			//先获得那个特定的聊天室
			roomID := client.RoomID
			// 如果那个房间的客户端集合不存在就创建出来 , _代表值, ok 代表这个值是否存在
			if _, ok := pool.Rooms[roomID]; !ok {
				pool.Rooms[roomID] = make(map[*Client]bool)
			}
			pool.Rooms[roomID][client] = true        //加入成功
			for client := range pool.Rooms[roomID] { //client 代表所有存在的Client
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("有人加入了聊天室")); err != nil {
					err := client.Conn.Close()
					if err != nil {
						fmt.Println(err)
						pool.Mu.Unlock() //*** return 之前记得把锁释放掉，不能重复占用锁资源
						return
					}
					fmt.Println(err)
					pool.Mu.Unlock()
					return
				}
			}
			fmt.Println(pool.Clients[client])
			pool.Mu.Unlock()
		case client := <-pool.Unregister:
			pool.Mu.Lock()
			// 先检查这个客户端集合是否存在
			roomID := client.RoomID
			fmt.Println(pool.Rooms[roomID])
			if ClientsCollection, ok := pool.Rooms[roomID]; ok {
				delete(ClientsCollection, client)
			} else {
				fmt.Println("客户端集合不存在")
				pool.Mu.Unlock()
				return
			}
			delete(pool.Clients, client) // 这个Clients代表所有的Client
			fmt.Println("客户端的数量", len(pool.Clients))
			client.RoomID = ""                       // 现在就没有加入房间，退出之后
			for client := range pool.Rooms[roomID] { // 遍历 key         value
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte("有人离开了聊天室"))
				if err != nil {
					err := client.Conn.Close()
					if err != nil {
						fmt.Println("send wrong", err)
						pool.Mu.Unlock()
						return
					}
					fmt.Println(err)
					pool.Mu.Unlock()
					return
				}
			}
			pool.Mu.Unlock()
			//不能每一个聊天室都发送消息
		case message := <-pool.Broadcast:
			pool.Mu.Lock()
			roomID := message.RoomId
			if _, ok := pool.Rooms[roomID]; !ok {
				log.Printf("Room not found: %s", message.RoomId)
				pool.Mu.Unlock()
				return
			}
			for client := range pool.Rooms[roomID] {
				if err := client.Conn.WriteJSON(message); err != nil {
					if err := client.Conn.Close(); err != nil {
						fmt.Println(err)
						pool.Mu.Unlock()
						return
					}
					fmt.Println(err)
					pool.Mu.Unlock()
					return
				}
			}
			pool.Mu.Unlock()
		}
	}
}
