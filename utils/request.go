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
	UserId    string `json:"user_id"`
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
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
}
