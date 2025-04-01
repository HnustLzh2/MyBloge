package db

import (
	"MyBloge/model"
	"MyBloge/utils"
	"github.com/google/uuid"
	"time"
)

func CreateChatRoomDB(name string, id string) (model.ChatRoom, error) {
	var room model.ChatRoom
	now := time.Now()
	room.ChatRoomId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	user, err := FindUserByID(id)
	if err != nil {
		return model.ChatRoom{}, err
	}
	room.RoomName = name
	room.RoomCreatorId = user.UserId
	room.RoomCreatorName = user.Name
	room.IsPrivate = false
	room.UserCollection = []model.User{}
	room.UserCollection = append(room.UserCollection, user)
	if err := sqlDb.AutoMigrate(&model.ChatRoom{}); err != nil {
		return room, err
	}
	if err := sqlDb.Create(&room); err != nil {
		return room, sqlDb.Error
	}
	return room, nil
}
func GetAllChatRoom() ([]model.ChatRoom, error) {
	var rooms []model.ChatRoom
	if err := sqlDb.Where("is_private = false").Find(&rooms).Error; err != nil {
		return rooms, err
	}
	return rooms, nil
}

func GetRoomById(id string) (model.ChatRoom, error) {
	var room model.ChatRoom
	if err := sqlDb.Where("chat_room_id = ?", id).Preload("UserCollection").First(&room).Error; err != nil {
		return room, err
	}
	return room, nil
}

func AddUserToChatRoom(userId string, roomId string) error {
	// 查找用户
	user, err := FindUserByID(userId)
	if err != nil {
		return err
	}
	// 查找聊天室
	room, err := GetRoomById(roomId)
	if err != nil {
		return err
	}
	// 使用事务确保操作的原子性
	tx := sqlDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 检查用户是否已经在聊天室中
	var count int64
	err = tx.Model(&model.ChatRoom{}).Joins("JOIN chat_room_users ON chat_rooms.chat_room_id = chat_room_users.chat_room_chat_room_id").
		Where("chat_room_users.user_user_id = ?", userId).
		Where("chat_rooms.chat_room_id = ?", roomId).
		Count(&count).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	if count > 0 {
		tx.Rollback()
		return nil // 用户已存在，直接返回
	}
	// 如果用户不存在，则添加
	err = tx.Model(&room).Association("UserCollection").Append(&user)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
func LeaveOutChatRoom(userId string, roomId string) error {
	user, err := FindUserByID(userId)
	if err != nil {
		return err
	}
	room, err := GetRoomById(roomId)
	if err != nil {
		return err
	}
	err = sqlDb.Model(&room).Association("UserCollection").Delete(&user)
	if err != nil {
		return err
	}
	return nil
}

func DeleteChatRoomDb(roomId string) error {
	if err := sqlDb.Delete(&model.ChatRoom{}, "chat_room_id = ?", roomId).Error; err != nil {
		return err
	}
	return nil
}

func AddMessageToDb(message model.Message) error {
	if err := sqlDb.AutoMigrate(&model.Message{}); err != nil {
		return err
	}
	if err := sqlDb.Create(&message).Error; err != nil {
		return err
	}
	return nil
}

func GetRoomMessage(roomId string) ([]model.Message, error) {
	var messages []model.Message
	// 获得这个房间的聊天记录
	if err := sqlDb.Where("room_id = ?", roomId).Find(&messages).Error; err != nil {
		return messages, err
	}
	return messages, nil
}

func CreatePrivateChatRoomDB(request utils.CreatePrivateRoomRequest) (model.ChatRoom, error) {
	var room model.ChatRoom
	now := time.Now()
	room.ChatRoomId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	userA, err := FindUserByID(request.UserAId)
	if err != nil {
		return room, err
	}
	userB, err := FindUserByID(request.UserBId)
	if err != nil {
		return room, err
	}
	room.UserCollection = []model.User{}
	room.UserCollection = append(room.UserCollection, userA, userB)
	room.RoomCreatorId = ""
	room.RoomCreatorName = request.RoomName
	room.IsPrivate = true
	if err := sqlDb.Create(&room).Error; err != nil {
		return model.ChatRoom{}, err
	}
	return room, nil
}

// FindUserInRoom 检查这个房间里这2个是不是已经存在了，存在了就不要重复添加了
func FindUserInRoom(idA string, idB string, roomId string) (bool, error) {
	var room model.ChatRoom
	room, err := GetRoomById(roomId)
	if err != nil {
		return false, err
	}
	var exists1 bool
	var exists2 bool
	sqlDb.Model(&room).Joins("UserCollection").Where("user_collection.user_id = ?", idA).First(&exists1)
	sqlDb.Model(&room).Joins("UserCollection").Where("user_collection.user_id = ?", idB).First(&exists2)
	if exists1 && exists2 {
		return true, nil
	}
	return false, nil
}

func GetAllPrivateRoom() ([]model.ChatRoom, error) {
	var rooms []model.ChatRoom
	if err := sqlDb.Where("is_private = true").Find(&rooms).Error; err != nil {
		return rooms, err
	}
	return rooms, nil
}

// GetYourRoomsDb 使用了自然连接，joins函数用于连接2个表在相同的部分，有左连接右连接，自然连接
func GetYourRoomsDb(userId string) ([]model.ChatRoom, error) {
	var rooms []model.ChatRoom
	// 使用 Joins 来关联 chat_room_users 中间表
	if err := sqlDb.Joins("JOIN chat_room_users ON chat_rooms.chat_room_id = chat_room_users.chat_room_chat_room_id").
		Where("chat_room_users.user_user_id = ?", userId).
		Preload("UserCollection").
		Find(&rooms).Error; err != nil {
		return rooms, err
	}
	return rooms, nil
}
