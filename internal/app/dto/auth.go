package dto

// RegisterRequest — тело запроса регистрации
type RegisterRequest struct {
	Login    string `json:"login"    example:"testuser"`
	Password string `json:"password" example:"secret"`
}

// LoginRequest — тело запроса входа
type LoginRequest struct {
	Login    string `json:"login"    example:"testuser"`
	Password string `json:"password" example:"secret"`
}
