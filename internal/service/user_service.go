package service

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/achemerzaev/url-shortening-service/internal/authorization"
	"github.com/achemerzaev/url-shortening-service/internal/models"
	"github.com/achemerzaev/url-shortening-service/internal/redisrepo"
	"github.com/achemerzaev/url-shortening-service/internal/repository"
	"github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"

	"context"
	"strconv"
	"strings"
	"time"
)

type UserService struct {
	repo      *repository.UserRepository
	redisrepo *redisrepo.RedisRepository
	logger    logger.Logger
}

func NewUserService(r *repository.UserRepository, redisr *redisrepo.RedisRepository, logger logger.Logger) *UserService {
	return &UserService{repo: r, redisrepo: redisr, logger: logger}
}

func (s *UserService) ServiceRegister(newUser models.PostUserRegistration) (models.User, models.Tokens, *errors.AppError) {
	var insertedUser models.User
	var tokens models.Tokens

	hash, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	err := bcrypt.CompareHashAndPassword(hash, []byte(newUser.Password))
	if err != nil {
		return insertedUser, tokens, errors.Wrap("INTERNAL_ERROR", "error creating user", err)
	}

	newUser.Password = string(hash)
	insertedUser, err = s.repo.RepoInsertUser(newUser)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return insertedUser, tokens, errors.ErrEmailExists
		}
		return insertedUser, tokens, errors.Wrap("INTERNAL_ERROR", "error creating user", err)
	}

	tokens.AccessToken, err = authorization.GenerateJWT(insertedUser.Id, 1*time.Hour)
	if err != nil {
		return insertedUser, tokens, errors.ErrGeneratingJWT
	}

	tokens.RefreshToken, err = authorization.GenerateJWT(insertedUser.Id, 7*24*time.Hour)
	if err != nil {
		return insertedUser, tokens, errors.ErrGeneratingJWT
	}

	err = s.redisrepo.SaveRefreshToken(context.Background(),
		strconv.Itoa(insertedUser.Id), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return insertedUser, tokens, errors.Wrap("INTERNAL_ERROR", "error creating user", err)
	}

	return insertedUser, tokens, nil
}

func (s *UserService) ServiceLogin(userinfo models.User) (models.Tokens, *errors.AppError) {
	var tokens models.Tokens
	storedPassword, err := s.repo.RepoRetrieveUser(userinfo.Email)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return tokens, errors.ErrInvalidCredentials
		}
		return tokens, errors.Wrap("INTERNAL_ERROR", "error logging in", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(userinfo.Password))
	if err != nil {
		return tokens, errors.ErrInvalidCredentials
	}

	tokens.AccessToken, err = authorization.GenerateJWT(userinfo.Id, 1*time.Hour)
	if err != nil {
		return tokens, errors.ErrGeneratingJWT
	}

	tokens.RefreshToken, err = authorization.GenerateJWT(userinfo.Id, 7*24*time.Hour)
	if err != nil {
		return tokens, errors.ErrGeneratingJWT
	}

	err = s.redisrepo.SaveRefreshToken(context.Background(),
		strconv.Itoa(userinfo.Id), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return tokens, errors.Wrap("INTERNAL_ERROR", "error logging in", err)
	}

	return tokens, nil
}

func (s *UserService) ServiceRefresh(refreshToken string) (models.Tokens, *errors.AppError) {
	var tokens models.Tokens

	userID, err := authorization.ValidateJWT(refreshToken)
	if err != nil {
		return tokens, errors.ErrInvalidToken
	}

	oldRefresh, err := s.redisrepo.GetRefreshToken(context.Background(), strconv.Itoa(userID))
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return tokens, errors.ErrInvalidToken
		}
		return tokens, errors.Wrap("INTERNAL_ERROR", "error refreshing token", err)
	}

	if oldRefresh != refreshToken {
		return tokens, errors.ErrInvalidToken
	}

	tokens.AccessToken, err = authorization.GenerateJWT(userID, 1*time.Hour)
	if err != nil {
		return tokens, errors.ErrGeneratingJWT
	}
	tokens.RefreshToken, err = authorization.GenerateJWT(userID, 7*24*time.Hour)
	if err != nil {
		return tokens, errors.ErrGeneratingJWT
	}

	err = s.redisrepo.SaveRefreshToken(context.Background(),
		strconv.Itoa(userID), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return tokens, errors.Wrap("INTERNAL_ERROR", "error refreshing token", err)
	}

	return tokens, nil
}
