//go:build integration

package service

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/achemerzaev/url-shortening-service/database"
	"github.com/achemerzaev/url-shortening-service/internal/config"
	"github.com/achemerzaev/url-shortening-service/internal/models"
	"github.com/achemerzaev/url-shortening-service/internal/redisrepo"
	"github.com/achemerzaev/url-shortening-service/internal/repository"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type integrationServices struct {
	userService   *UserService
	urlService    *URLService
	userRepo      *repository.UserRepository
	urlRepo       *repository.URLRepository
	redisUserRepo *redisrepo.RedisUserRepository
	redisURLRepo  *redisrepo.RedisURLRepository
}

func setupIntegrationServices(t *testing.T) *integrationServices {
	t.Helper()

	envPath := filepath.Join("..", "..", ".env.test")
	require.NoError(t, godotenv.Load(envPath))

	dbCfg := config.DBConfig{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Name:     os.Getenv("POSTGRES_DB"),
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	require.NoError(t, err)

	redisCfg := config.RedisConfig{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   redisDB,
	}

	pool, err := database.NewPostgres(dbCfg)
	require.NoError(t, err)

	redisClient, err := database.NewRedis(redisCfg)
	require.NoError(t, err)

	log, err := logger.New("debug")
	require.NoError(t, err)

	cleanupIntegrationStorage(t, pool, redisClient)
	t.Cleanup(func() {
		cleanupIntegrationStorage(t, pool, redisClient)
		_ = redisClient.Close()
		pool.Close()
	})

	userRepo := repository.NewUserRepository(pool)
	urlRepo := repository.NewURLRepository(pool)
	redisUserRepo := redisrepo.NewRedisUserRepository(redisClient, log)
	redisURLRepo := redisrepo.NewRedisURLRepository(redisClient, log)

	return &integrationServices{
		userService:   NewUserService(userRepo, redisUserRepo),
		urlService:    NewUrlService(urlRepo, redisURLRepo),
		userRepo:      userRepo,
		urlRepo:       urlRepo,
		redisUserRepo: redisUserRepo,
		redisURLRepo:  redisURLRepo,
	}
}

func cleanupIntegrationStorage(t *testing.T, pool *pgxpool.Pool, redisClient *redis.Client) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, "TRUNCATE TABLE urls, url_users RESTART IDENTITY CASCADE")
	require.NoError(t, err)
	require.NoError(t, redisClient.FlushDB(ctx).Err())
}

func createIntegrationUser(t *testing.T, svc *UserService, email string) models.User {
	t.Helper()

	user, _, err := svc.ServiceRegister(context.Background(), models.User{
		Name:     "test-user",
		Email:    email,
		Password: "secret-pass",
	})
	require.NoError(t, err)

	return user
}

func TestUserServiceIntegration_RegisterLoginRefreshFlow(t *testing.T) {
	svc := setupIntegrationServices(t)
	ctx := context.Background()

	registeredUser, registerTokens, err := svc.userService.ServiceRegister(ctx, models.User{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "super-secret",
	})
	require.NoError(t, err)
	require.NotZero(t, registeredUser.Id)
	require.NotEmpty(t, registerTokens.AccessToken)
	require.NotEmpty(t, registerTokens.RefreshToken)

	retrievedUser, err := svc.userRepo.RepoRetrieveUser(ctx, registeredUser.Email)
	require.NoError(t, err)
	require.NotEqual(t, "super-secret", retrievedUser.Password)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(retrievedUser.Password), []byte("super-secret")))

	storedRefresh, err := svc.redisUserRepo.GetRefreshToken(ctx, strconv.Itoa(registeredUser.Id))
	require.NoError(t, err)
	require.Equal(t, registerTokens.RefreshToken, storedRefresh)

	loginTokens, err := svc.userService.ServiceLogin(ctx, models.User{
		Id:       registeredUser.Id,
		Email:    registeredUser.Email,
		Password: "super-secret",
	})
	require.NoError(t, err)
	require.NotEmpty(t, loginTokens.AccessToken)
	require.NotEmpty(t, loginTokens.RefreshToken)
	require.NotEqual(t, registerTokens.RefreshToken, loginTokens.RefreshToken)

	storedRefresh, err = svc.redisUserRepo.GetRefreshToken(ctx, strconv.Itoa(registeredUser.Id))
	require.NoError(t, err)
	require.Equal(t, loginTokens.RefreshToken, storedRefresh)

	refreshedTokens, err := svc.userService.ServiceRefresh(ctx, loginTokens.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, refreshedTokens.AccessToken)
	require.NotEmpty(t, refreshedTokens.RefreshToken)
	require.NotEqual(t, loginTokens.RefreshToken, refreshedTokens.RefreshToken)

	storedRefresh, err = svc.redisUserRepo.GetRefreshToken(ctx, strconv.Itoa(registeredUser.Id))
	require.NoError(t, err)
	require.Equal(t, refreshedTokens.RefreshToken, storedRefresh)
}

func TestUserServiceIntegration_ErrorCases(t *testing.T) {
	svc := setupIntegrationServices(t)
	ctx := context.Background()

	_, _, err := svc.userService.ServiceRegister(ctx, models.User{
		Name:     "Alice",
		Email:    "duplicate@example.com",
		Password: "super-secret",
	})
	require.NoError(t, err)

	_, _, err = svc.userService.ServiceRegister(ctx, models.User{
		Name:     "Alice 2",
		Email:    "duplicate@example.com",
		Password: "another-secret",
	})
	require.ErrorIs(t, err, appErr.ErrEmailExists)

	_, err = svc.userService.ServiceLogin(ctx, models.User{
		Id:       999,
		Email:    "missing@example.com",
		Password: "super-secret",
	})
	require.ErrorIs(t, err, appErr.ErrInvalidCredentials)

	user := createIntegrationUser(t, svc.userService, "login@example.com")
	_, err = svc.userService.ServiceLogin(ctx, models.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: "wrong-password",
	})
	require.ErrorIs(t, err, appErr.ErrInvalidCredentials)

	_, err = svc.userService.ServiceRefresh(ctx, "invalid.refresh.token")
	require.ErrorIs(t, err, appErr.ErrInvalidToken)
}

func TestURLServiceIntegration_CRUDAndCacheFlow(t *testing.T) {
	svc := setupIntegrationServices(t)
	ctx := context.Background()
	owner := createIntegrationUser(t, svc.userService, "owner@example.com")

	createdURL, err := svc.urlService.ServicePost(ctx, models.URLInfo{
		Url:     "example.com",
		OwnerID: owner.Id,
	})
	require.NoError(t, err)
	require.NotZero(t, createdURL.Id)
	require.Equal(t, "https://example.com", createdURL.Url)
	require.Len(t, createdURL.ShortCode, 6)

	fromPostgres, err := svc.urlRepo.RepositoryGet(ctx, createdURL.ShortCode)
	require.NoError(t, err)
	require.Equal(t, createdURL.Url, fromPostgres.Url)
	require.Equal(t, 0, fromPostgres.AccessCount)

	longURL, err := svc.urlService.ServiceGet(ctx, createdURL.ShortCode, owner.Id)
	require.NoError(t, err)
	require.Equal(t, createdURL.Url, longURL)

	cachedStats, err := svc.redisURLRepo.GetUrlStats(ctx, createdURL.ShortCode, owner.Id)
	require.NoError(t, err)
	require.Equal(t, 1, cachedStats.AccessCount)

	longURL, err = svc.urlService.ServiceGet(ctx, createdURL.ShortCode, owner.Id)
	require.NoError(t, err)
	require.Equal(t, createdURL.Url, longURL)

	stats, err := svc.urlService.ServiceGetStats(ctx, createdURL.ShortCode, owner.Id)
	require.NoError(t, err)
	require.Equal(t, 2, stats.AccessCount)

	updatedURL, err := svc.urlService.ServicePut(ctx, createdURL.ShortCode, "golang.org", owner.Id)
	require.NoError(t, err)
	require.Equal(t, "https://golang.org", updatedURL.Url)

	require.NoError(t, svc.urlService.ServiceDelete(ctx, createdURL.ShortCode, owner.Id))

	_, err = svc.urlRepo.RepositoryGet(ctx, createdURL.ShortCode)
	require.ErrorIs(t, err, appErr.ErrNotFound)

	_, err = svc.redisURLRepo.GetUrlStats(ctx, createdURL.ShortCode, owner.Id)
	require.Error(t, err)
}

func TestURLServiceIntegration_ErrorCases(t *testing.T) {
	svc := setupIntegrationServices(t)
	ctx := context.Background()

	owner := createIntegrationUser(t, svc.userService, "owner-errors@example.com")
	otherUser := createIntegrationUser(t, svc.userService, "other-errors@example.com")

	createdURL, err := svc.urlService.ServicePost(ctx, models.URLInfo{
		Url:     "forbidden.example.com",
		OwnerID: owner.Id,
	})
	require.NoError(t, err)

	_, err = svc.urlService.ServiceGetStats(ctx, createdURL.ShortCode, owner.Id)
	require.NoError(t, err)

	_, err = svc.urlService.ServiceGet(ctx, createdURL.ShortCode, otherUser.Id)
	require.ErrorIs(t, err, appErr.ErrForbidden)

	_, err = svc.urlService.ServicePut(ctx, createdURL.ShortCode, "new.example.com", otherUser.Id)
	require.ErrorIs(t, err, appErr.ErrForbidden)

	err = svc.urlService.ServiceDelete(ctx, createdURL.ShortCode, otherUser.Id)
	require.ErrorIs(t, err, appErr.ErrForbidden)

	_, err = svc.urlService.ServiceGet(ctx, "missing1", owner.Id)
	require.ErrorIs(t, err, appErr.ErrNotFound)

	_, err = svc.urlService.ServicePut(ctx, "missing2", "still-missing.example.com", owner.Id)
	require.ErrorIs(t, err, appErr.ErrNotFound)

	err = svc.urlService.ServiceDelete(ctx, "missing3", owner.Id)
	require.ErrorIs(t, err, appErr.ErrNotFound)

	_, err = svc.urlService.ServiceGetStats(ctx, "missing4", owner.Id)
	require.ErrorIs(t, err, appErr.ErrNotFound)
}
