package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/achemerzaev/url-shortening-service/internal/handler"
	"github.com/achemerzaev/url-shortening-service/internal/redisrepo"
	"github.com/achemerzaev/url-shortening-service/internal/repository"
	"github.com/achemerzaev/url-shortening-service/internal/service"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"
)

type Handlers struct {
	URL  *handler.URLHandler
	User *handler.UserHandler
}

func BuildHandlers(pgPool *pgxpool.Pool, rdb *redis.Client, logger logger.Logger) Handlers {
	// Redis
	redisURLRepo := redisrepo.NewRedisURLRepository(rdb, logger)
	redisUserRepo := redisrepo.NewRedisUserRepository(rdb, logger)

	// URL
	urlRepo := repository.NewURLRepository(pgPool)
	urlService := service.NewUrlService(urlRepo, redisURLRepo)
	urlHandler := handler.NewUrlHandler(urlService, logger)

	// User
	userRepo := repository.NewUserRepository(pgPool)
	userService := service.NewUserService(userRepo, redisUserRepo)
	userHandler := handler.NewUserHandler(userService, logger)

	return Handlers{
		URL:  urlHandler,
		User: userHandler,
	}
}
