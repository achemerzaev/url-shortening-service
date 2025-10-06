package service

import (
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/boretsotets/url-shortening-service/internal/authorization"
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/repository"

	"context"
	"errors"
	"strconv"
	"time"
)

type UserService struct {
	repo      *repository.UserRepository
	redisrepo *redisrepo.RedisRepository
	logger    *zap.Logger
}

func NewUserService(r *repository.UserRepository, redisr *redisrepo.RedisRepository, logger *zap.Logger) *UserService {
	return &UserService{repo: r, redisrepo: redisr, logger: logger}
}

func (s *UserService) ServiceRegister(newUser models.PostUserRegistration) (models.User, models.Tokens, error) {
	var insertedUser models.User
	var tokens models.Tokens

	hash, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	err := bcrypt.CompareHashAndPassword(hash, []byte(newUser.Password))
	if err != nil {
		return insertedUser, tokens, err
	}

	newUser.Password = string(hash)
	insertedUser, err = s.repo.RepoInsertUser(newUser)
	if err != nil {
		return insertedUser, tokens, err
	}

	tokens.AccessToken, err = authorization.GenerateJWT(insertedUser.Id, 1*time.Hour)
	if err != nil {
		return insertedUser, tokens, err
	}

	tokens.RefreshToken, err = authorization.GenerateJWT(insertedUser.Id, 7*24*time.Hour)
	if err != nil {
		return insertedUser, tokens, err
	}

	err = s.redisrepo.SaveRefreshToken(context.Background(),
		strconv.Itoa(insertedUser.Id), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return insertedUser, tokens, err
	}

	return insertedUser, tokens, nil
}

func (s *UserService) ServiceLogin(userinfo models.User) (models.Tokens, error) {
	var tokens models.Tokens
	storedPassword, err := s.repo.RepoRetrieveUser(userinfo.Email)
	if err != nil {
		return tokens, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(userinfo.Password))
	if err != nil {
		return tokens, err
	}

	tokens.AccessToken, err = authorization.GenerateJWT(userinfo.Id, 1*time.Hour)
	if err != nil {
		return tokens, err
	}

	tokens.RefreshToken, err = authorization.GenerateJWT(userinfo.Id, 7*24*time.Hour)
	if err != nil {
		return tokens, err
	}

	err = s.redisrepo.SaveRefreshToken(context.Background(),
		strconv.Itoa(userinfo.Id), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		return tokens, err
	}

	return tokens, nil
}

func (s *UserService) ServiceRefresh(userID int, refreshtoken string) (models.Tokens, error) {
	var tokens models.Tokens
	oldRefresh, err := s.redisrepo.GetRefreshToken(context.Background(), strconv.Itoa(userID))
	if err != nil {
		return tokens, err
	}
	s.logger.Info("good so far", zap.String("old refresh: ", oldRefresh))
	s.logger.Info("good so far", zap.String("inserted refresh: ", refreshtoken))

	if oldRefresh != refreshtoken {
		return tokens, errors.New("refresh token is not valid")
	}
	s.logger.Info("are equal")

	tokens.AccessToken, err = authorization.GenerateJWT(userID, 1*time.Hour)
	if err != nil {
		s.logger.Info("error generating access token")
		return tokens, err
	}
	tokens.RefreshToken, err = authorization.GenerateJWT(userID, 7*24*time.Hour)
	if err != nil {
		s.logger.Info("error generating refresh token")
		return tokens, err
	}

	err = s.redisrepo.SaveRefreshToken(context.Background(),
		strconv.Itoa(userID), tokens.RefreshToken, 7*24*time.Hour)
	if err != nil {
		s.logger.Info("error in saving")
		return tokens, err
	}
	return tokens, err

}
