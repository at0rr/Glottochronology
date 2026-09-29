package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"glottochronology/internal/app/ds"
)

// POST /api/auth/register
func (h *Handler) APIRegister(c *gin.Context) {
	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("неверное тело запроса"))
		return
	}
	if body.Login == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, ds.NewError("login и password обязательны"))
		return
	}

	existing, _ := h.Repository.GetUserByLogin(body.Login)
	if existing != nil {
		c.JSON(http.StatusConflict, ds.NewError("пользователь уже существует"))
		return
	}

	u := &ds.User{Login: body.Login, Password: body.Password}
	if err := h.Repository.CreateUser(u); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, ds.Success(ds.NewUserSerializer(*u)))
}

// POST /api/auth/login — заглушка для 4 лабораторной
func (h *Handler) APILogin(c *gin.Context) {
	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("неверное тело"))
		return
	}

	u, _ := h.Repository.GetUserByLogin(body.Login)
	if u == nil || u.Password != body.Password {
		c.JSON(http.StatusUnauthorized, ds.NewError("неверные учётные данные"))
		return
	}

	c.JSON(http.StatusOK, ds.Success(map[string]interface{}{
		"message": "заглушка: аутентификация будет в 4 лабе",
		"user":    ds.NewUserSerializer(*u),
		"token":   "stub-token-" + body.Login,
	}))
}

// POST /api/auth/logout — заглушка
func (h *Handler) APILogout(c *gin.Context) {
	c.JSON(http.StatusOK, ds.Success(map[string]string{
		"message": "заглушка: деавторизация будет в 4 лабе",
	}))
}
