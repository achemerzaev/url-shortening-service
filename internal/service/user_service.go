package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/achemerzaev/url-shortening-service/internal/authorization"
	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"context"
	"errors"
	"strconv"
	"time"
)

type UserRepository interface {
	RepoInsertUser(ctx context.Context, newUser models.PostUserRegistration) (models.User, error)
	RepoRetrieveUser(ctx context.Context, email string) (string, error)
}

type RedisUserRepository interface {
	SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
}

type UserService struct {
	repo      UserRepository
	redisrepo RedisUserRepository
	logger    logger.Logger
}

func NewUserService(r UserRepository, redisr RedisUserRepository, logger logger.Logger) *UserService {
	return &UserService{repo: r, redisrepo: redisr, logger: logger}
}

func (s *UserService) ServiceRegister(ctx context.Context, newUser models.PostUserRegistration) (models.User, models.Tokens, error) {
	var insertedUser models.User
	var tokens models.Tokens

	hash, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	err := bcrypt.CompareHashAndPassword(hash, []byte(newUser.Password))
	if err != nil {
		return insertedUser, tokens, fmt.Errorf("service register error hashing password: %w", err)
	}

	newUser.Password = string(hash)
	insertedUser, err = s.repo.RepoInsertUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, appErr.ErrEmailExists) {
			return insertedUser, tokens, appErr.ErrEmailExists
		}
		return insertedUser, tokens, fmt.Errorf("service register error repo inserting user: %w", err)
	}

	tokens.AccessToken, err = authorization.GenerateJWT(insertedUser.Id, 1*time.Hour)
	if err != nil {
		return insertedUser, tokens, appErr.ErrGeneratingJWT
	}

	tokens.RefreshToken, err = authorization.GenerateJWT(insertedUser.Id, 7*24*time.Hour)
	if err != nil {
		return insertedUser, tokens, appErr.ErrGeneratingJWT
	}

	err = s.redisrepo.SaveRefreshToken(ctx,
		strconv.Itoa(insertedUser.Id), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return insertedUser, tokens, fmt.Errorf("service register error saving token to redis: %w", err)
	}

	return insertedUser, tokens, nil
}

func (s *UserService) ServiceLogin(ctx context.Context, userinfo models.User) (models.Tokens, error) {
	var tokens models.Tokens
	storedPassword, err := s.repo.RepoRetrieveUser(ctx, userinfo.Email)
	if err != nil {
		if errors.Is(err, appErr.ErrInvalidCredentials) {
			return tokens, appErr.ErrInvalidCredentials
		}
		return tokens, fmt.Errorf("service login repo retrieve user error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(userinfo.Password))
	if err != nil {
		return tokens, appErr.ErrInvalidCredentials
	}

	tokens.AccessToken, err = authorization.GenerateJWT(userinfo.Id, 1*time.Hour)
	if err != nil {
		return tokens, appErr.ErrGeneratingJWT
	}

	tokens.RefreshToken, err = authorization.GenerateJWT(userinfo.Id, 7*24*time.Hour)
	if err != nil {
		return tokens, appErr.ErrGeneratingJWT
	}

	err = s.redisrepo.SaveRefreshToken(ctx,
		strconv.Itoa(userinfo.Id), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return tokens, fmt.Errorf("service login save refresh token err: %w", err)
	}

	return tokens, nil
}

func (s *UserService) ServiceRefresh(ctx context.Context, refreshToken string) (models.Tokens, error) {
	var tokens models.Tokens

	userID, err := authorization.ValidateJWT(refreshToken)
	if err != nil {
		return tokens, appErr.ErrInvalidToken
	}

	oldRefresh, err := s.redisrepo.GetRefreshToken(ctx, strconv.Itoa(userID))
	if err != nil {
		return tokens, fmt.Errorf("service refresh get refresh token err: %w", err)
	}

	if oldRefresh != refreshToken {
		return tokens, appErr.ErrInvalidToken
	}

	tokens.AccessToken, err = authorization.GenerateJWT(userID, 1*time.Hour)
	if err != nil {
		return tokens, appErr.ErrGeneratingJWT
	}
	tokens.RefreshToken, err = authorization.GenerateJWT(userID, 7*24*time.Hour)
	if err != nil {
		return tokens, appErr.ErrGeneratingJWT
	}

	err = s.redisrepo.SaveRefreshToken(ctx,
		strconv.Itoa(userID), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return tokens, fmt.Errorf("service refresh save refresh token err: %w", err)
	}

	return tokens, nil
}
