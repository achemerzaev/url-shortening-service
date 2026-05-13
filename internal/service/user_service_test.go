package service

import (
	"context"
	"testing"
	"time"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	db     map[string]models.User
	LastId int
}

type mockRedisUserRepository struct {
	db map[string]RedisRow
}

type RedisRow struct {
	Token string
	Ttl   time.Duration
}

func (r *mockUserRepository) RepoInsertUser(ctx context.Context, newUser models.User) (models.User, error) {
	if _, ok := r.db[newUser.Email]; ok {
		return models.User{}, appErr.ErrEmailExists
	}
	r.LastId += 1
	insert := models.User{
		Id:       r.LastId,
		Name:     newUser.Name,
		Email:    newUser.Email,
		Password: newUser.Password,
	}
	r.db[insert.Email] = insert
	return insert, nil
}

func (r *mockUserRepository) RepoRetrieveUser(ctx context.Context, email string) (models.User, error) {
	if v, ok := r.db[email]; !ok {
		return v, appErr.ErrNotFound
	} else {
		return v, nil
	}
}

func (r *mockRedisUserRepository) SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	r.db[userID] = RedisRow{
		Token: token,
		Ttl:   ttl,
	}
	return nil
}

func (r *mockRedisUserRepository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	if v, ok := r.db[userID]; !ok {
		return "", appErr.ErrNotFound
	} else {
		return v.Token, nil
	}
}

func setupusersvc() *UserService {
	repo := &mockUserRepository{
		db: make(map[string]models.User),
	}

	redisrepo := &mockRedisUserRepository{
		db: make(map[string]RedisRow),
	}

	return NewUserService(repo, redisrepo)
}

func TestServiceRegister_Success(t *testing.T) {
	svc := setupusersvc()

	newUser := models.User{
		Name:     "John",
		Email:    "john@gmail.com",
		Password: "john123",
	}

	user, _, err := svc.ServiceRegister(context.Background(), newUser)

	require.NoError(t, err)
	require.Equal(t, newUser.Name, user.Name)
	require.Equal(t, newUser.Email, user.Email)
}

func TestServiceLogin_Success(t *testing.T) {
	svc := setupusersvc()

	newUser := models.User{
		Name:     "John",
		Email:    "john@gmail.com",
		Password: "john123",
	}

	user, _, _ := svc.ServiceRegister(context.Background(), newUser)
	user.Password = newUser.Password
	_, err := svc.ServiceLogin(context.Background(), user)

	require.NoError(t, err)
}

func TestServiceLogin_NotFound(t *testing.T) {
	svc := setupusersvc()

	newUser := models.User{
		Name:     "John",
		Email:    "john@gmail.com",
		Password: "john123",
	}

	user, _, _ := svc.ServiceRegister(context.Background(), newUser)

	user.Email = "johny@gmail.com"
	_, err := svc.ServiceLogin(context.Background(), user)

	require.ErrorIs(t, err, appErr.ErrNotFound)
}

func TestServiceLogin_InvalidPassword(t *testing.T) {
	svc := setupusersvc()

	newUser := models.User{
		Name:     "John",
		Email:    "john@gmail.com",
		Password: "john123",
	}

	user, _, _ := svc.ServiceRegister(context.Background(), newUser)

	user.Password = "john321"
	_, err := svc.ServiceLogin(context.Background(), user)

	require.ErrorIs(t, err, appErr.ErrInvalidCredentials)
}

func TestServiceRefresh_Success(t *testing.T) {
	svc := setupusersvc()

	newUser := models.User{
		Name:     "John",
		Email:    "john@gmail.com",
		Password: "john123",
	}

	_, tokens, _ := svc.ServiceRegister(context.Background(), newUser)

	_, err := svc.ServiceRefresh(context.Background(), tokens.RefreshToken)

	require.NoError(t, err)
}

func TestServiceRefresh_InvalidRefreshToken(t *testing.T) {
	svc := setupusersvc()

	var tokens models.Tokens
	tokens.RefreshToken = "invalid.refresh.token"
	_, err := svc.ServiceRefresh(context.Background(), tokens.RefreshToken)

	require.ErrorIs(t, err, appErr.ErrInvalidToken)
}
