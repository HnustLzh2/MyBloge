package model

const (
	MessageTypeText   = 1 // 文本消息
	MessageTypeBinary = 2 // 二进制消息
)

// Message type 1是文本类型 2是二进制消息
type Message struct {
	MessageId   string `json:"message_id" gorm:"primaryKey;"`
	RoomId      string `json:"room_id" gorm:"foreignKey:RoomID;references:ChatRoomId"` // 关联 ChatRoom 的 ChatRoomId
	MessageType int    `json:"message_type"`
	Message     string `json:"message"`
	SenderId    string `json:"sender_id" gorm:"foreignKey:SenderId;references:UserId"`
	Timestamp   int64  `json:"timestamp"`
}

// ChatRoom UserCollection是多对多关系，设置中间表
type ChatRoom struct {
	ChatRoomId      string `json:"chat_room_id" gorm:"primaryKey;"`
	RoomName        string `json:"room_name"`
	IsPrivate       bool   `json:"is_private"`
	RoomCreatorId   string `json:"room_creator" gorm:"foreignKey:RoomCreatorId;references:UserId"` // 创建者信息
	RoomCreatorName string `json:"room_creator_name"`
	UserCollection  []User `json:"user_collection" gorm:"many2many:chat_room_users;default:[]User"`
}
