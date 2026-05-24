// Package authorization реализует функции для генерации и валидации
// JWT токена при авторизации пользователей
package authorization

import (
	"github.com/golang-jwt/jwt/v5"

	"crypto/rand"
	"encoding/hex"
	"time"
)

// Секретный пароль, использующийся для
// подписи JWT токенов
var jwtKey = []byte("bigsecret") // []byte(os.Getenv("JWT_SECRET"))

// GenerateJWT генерирует новый токен авторизации для пользователя.
// Включает идентификатор пользователя и время жизни токена - 1 час. В
// Используется при авторизации после входа в систему. Возвращает токен или ошибку

func GenerateJWT(userID int, ttl time.Duration) (string, error) {
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(ttl).Unix(),
		"jti":     hex.EncodeToString(jti),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateJWT проверяет, является ли tokenString валидным токеном
// авторизации пользователя. Если токен действителен, возвращает
// идентификатор пользователя userID. Если токен недействителен или
// истек, возвращает ошибку
func ValidateJWT(tokenString string) (int, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	return userID, nil
}
