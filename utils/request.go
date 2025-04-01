package utils

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	Username      string `json:"username"`
	Authorization string `json:"authorization"`
}
type AddArticleRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Preview  string `json:"preview"`
	Category string `json:"category"`
}
type FavoriteArticleRequest struct {
	ArticleId string `json:"article_id"`
	FolderId  string `json:"folder_id"`
}
type UserIdRequest struct {
	UserId string `json:"user_id"`
}
type ArticleIdRequest struct {
	ArticleId string `json:"article_id"`
}
type ModifyArticleRequest struct {
	ID       string `json:"id"`
	Title    string `json:"title" default:""`
	Content  string `json:"content" default:""`
	Preview  string `json:"preview" default:""`
	Category string `json:"category" default:""`
}
type AddCommentRequest struct {
	Content    string `json:"content"`
	SendUserId string `json:"user_id"`
	ArticleId  string `json:"article_id"`
}
type RepliedCommentRequest struct {
	Content    string `json:"content"`
	SendUserId string `json:"user_id"`
	ArticleId  string `json:"article_id"`
	ParentID   string `json:"parent_id"`
}
type LikeCommentRequest struct {
	CommentId string `json:"comment_id"`
	UserId    string `json:"user_id"`
}
type TokenRequest struct {
	AuthToken    string `json:"auth_token"`
	RefreshToken string `json:"refresh_token"`
}
type CreateCustomizeFolderRequest struct {
	FolderName string `json:"folder_name"`
	UserId     string `json:"user_id"`
}
type ModifyCustomizeFolderRequest struct {
	NewName  string `json:"new_name"`
	FolderId string `json:"folder_id"`
}
type CreateRoomRequest struct {
	RoomName  string `json:"room_name"`
	CreatorId string `json:"creator_id"`
}
type SendMessageRequest struct {
	MessageId      string `json:"message_id"`
	MessageContent string `json:"message_content"`
	MessageType    int    `json:"message_type"`
	RoomId         string `json:"room_id"`
	SenderId       string `json:"sender_id"`
	Timestamp      int64  `json:"timestamp"`
}
type CreatePrivateRoomRequest struct {
	RoomName string `json:"room_name"`
	UserAId  string `json:"user_a_id"`
	UserBId  string `json:"user_b_id"`
}
