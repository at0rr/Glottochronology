package handler

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"glottochronology/internal/app/ds"
)

// Register godoc
// @Summary      Регистрация пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      201  {object}  ds.UserSerializer
// @Failure      400  {object}  ds.ErrorResponse
// @Failure      409  {object}  ds.ErrorResponse
// @Router       /api/auth/register [post]
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

	u, err := h.Repository.Register(body.Login, body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, ds.Success(ds.NewUserSerializer(*u)))
}

// Login godoc
// @Summary      Вход в систему
// @Tags         auth
// @Accept       json
// @Produce      json
// @Failure      401  {object}  ds.ErrorResponse
// @Router       /api/auth/login [post]
func (h *Handler) APILogin(c *gin.Context) {
	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("неверное тело запроса"))
		return
	}

	u, _ := h.Repository.GetUserByLogin(body.Login)
	if u == nil || !h.Repository.CheckPassword(u.Password, body.Password) {
		c.JSON(http.StatusUnauthorized, ds.NewError("неверные учётные данные"))
		return
	}

	expiresIn, _ := strconv.Atoi(os.Getenv("JWT_EXPIRES_IN"))
	if expiresIn == 0 {
		expiresIn = 3600
	}

	claims := &ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "glotto",
		},
		UserID: u.ID,
		Login:  u.Login,
		Role:   u.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"expires_in":   expiresIn,
			"access_token": tokenStr,
			"token_type":   "Bearer",
			"user":         ds.NewUserSerializer(*u),
		},
	})
}

// Logout godoc
// @Summary      Выход (добавляет токен в blacklist)
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ds.SuccessResponse
// @Failure      401  {object}  ds.ErrorResponse
// @Router       /api/auth/logout [post]
func (h *Handler) APILogout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 8 {
		c.JSON(http.StatusUnauthorized, ds.NewError("нет токена"))
		return
	}
	tokenStr := authHeader[7:]

	expiresIn, _ := strconv.Atoi(os.Getenv("JWT_EXPIRES_IN"))
	if expiresIn == 0 {
		expiresIn = 3600
	}

	if err := h.Redis.WriteJWTToBlacklist(
		c.Request.Context(),
		tokenStr,
		time.Duration(expiresIn)*time.Second,
	); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ds.Success(map[string]string{"message": "вы вышли из системы"}))
}
