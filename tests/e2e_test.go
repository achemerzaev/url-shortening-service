package e2e

import (
	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/handler"
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/authorization"



	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"net/http"
	"net/http/httptest"
	"encoding/json"
	"testing"
	"context"
	"strings"
)

type TestApp struct {
	Router *gin.Engine
	UserRepo *repository.UserRepository
	UrlRepo *repository.UrlRepository
}

func setupTestApp(t *testing.T) *TestApp {
	t.Helper()

	ctx := context.Background()

	dsn := "postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
	pool, err := database.InitDb(ctx, dsn)
	require.NoError(t, err)

	redisClient, err := database.InitRedis("localhost:6379", "", 0)
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	redisRepo := redisrepo.NewRedisRepository(redisClient, logger)

	urlRepo := repository.NewUrlRepository(pool, logger)
	urlService := service.NewUrlService(urlRepo, redisRepo, logger)
	urlHandler := handler.NewUrlHandler(urlService, logger)

	userRepo := repository.NewUserRepository(pool, logger)
	userService := service.NewUserService(userRepo, redisRepo, logger)
	userHandler := handler.NewUserHandler(userService, logger)


	router := gin.Default()

	router.POST("/shorten", urlHandler.HandlerPost)
	router.GET("/shorten/:shortcode", urlHandler.HandlerGet)
	router.PUT("/shorten/:shortcode", urlHandler.HandlerPut)
	router.DELETE("/shorten/:shortcode", urlHandler.HandlerDelete)
	router.GET("/shorten/:shortcode/stats", urlHandler.HandlerGetStats)
	router.POST("/register", userHandler.HandlerRegister)
	router.POST("/login", userHandler.HandlerLogin)
	router.POST("/refresh", userHandler.HandlerRefresh)

	return &TestApp{
		Router: router,
		UserRepo: userRepo,
		UrlRepo: urlRepo,
	}
}

func TestTokenFunctionality(t *testing.T) {
	app := setupTestApp(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"a", "email":"bb", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)

	// проверка статуса
	require.Equal(t, http.StatusCreated, w.Code)

	// проверка, что токены были созданы
	var tokens models.Tokens
	json.NewDecoder(w.Body).Decode(&tokens)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)

	// проверка, что юзер был добавлен в бд
	checkUserCreation, err := app.UserRepo.RepoRetrieveUser("bb")
	require.NoError(t, err)
	require.NotEmpty(t, checkUserCreation)

	// проверка, что аксесс работает на другие хендлеры
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "vk.com"}`))
	req1.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code)

	// проверка, что рефреш токен работает на рефреш эндпоинт
	w2 := httptest.NewRecorder()
	refreshtokenrequest := `{"refreshtoken": "` + tokens.RefreshToken + `"}`
	req2 := httptest.NewRequest("POST", "/refresh", strings.NewReader(refreshtokenrequest))
	app.Router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusCreated, w2.Code)
}

func TestCrudOperations(t *testing.T) {
	app := setupTestApp(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"a", "email":"bbb", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)

	var tokens models.Tokens
	json.NewDecoder(w.Body).Decode(&tokens)

	// проверка создания короткой ссылки
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "mail.ru"}`))
	req1.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusOK, w1.Code)
	var data models.UrlInfo
	json.NewDecoder(w1.Body).Decode(&data)

	require.NotEmpty(t, data.Id)
	require.Equal(t, data.Url, "https://mail.ru")
	require.NotEmpty(t, data.ShortCode)
	require.NotEmpty(t, data.CreatedAt)
	require.NotEmpty(t, data.UpdatedAt)
	require.Equal(t, data.CreatedAt, data.UpdatedAt)
	require.Equal(t, data.AccessCount, 0)
	require.NotEmpty(t, data.OwnerID)
	tokenOwnerId, _ := authorization.ValidateJWT(tokens.AccessToken)
	require.Equal(t, data.OwnerID, tokenOwnerId)

	// проверка редиректа
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/shorten/"+data.ShortCode, nil)
	req2.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusFound, w2.Code)

	res := w2.Result()
	defer res.Body.Close()
	loc, _ := res.Location()
	require.Equal(t, "https://mail.ru", loc.String())

	// проверка получения статистики
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/shorten/"+data.ShortCode+"/stats", nil)
	req3.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w3, req3)

	var requestedData models.UrlInfo
	json.NewDecoder(w3.Body).Decode(&requestedData)

	require.Equal(t, http.StatusOK, w3.Code)
	require.Equal(t, data.Id, requestedData.Id)
	require.Equal(t, "https://mail.ru", requestedData.Url)
	require.Equal(t, data.ShortCode, requestedData.ShortCode)
	require.Equal(t, data.CreatedAt, requestedData.CreatedAt)
	require.Equal(t, data.UpdatedAt, requestedData.UpdatedAt)
	require.Equal(t, data.AccessCount+1, requestedData.AccessCount)
	require.Equal(t, data.OwnerID, requestedData.OwnerID)

	// проверка изменения задачи
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("PUT", "/shorten/"+data.ShortCode, strings.NewReader(`{"url": "go.dev"}`))
	req4.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w4, req4)	

	var changedData models.UrlInfo
	json.NewDecoder(w4.Body).Decode(&changedData)

	require.Equal(t, http.StatusOK, w4.Code)
	require.Equal(t, requestedData.Id, changedData.Id)
	require.NotEqual(t, requestedData.Url, changedData.Url)
	require.Equal(t, "https://go.dev", changedData.Url)
	require.Equal(t, requestedData.ShortCode, changedData.ShortCode)
	require.Equal(t, requestedData.CreatedAt, changedData.CreatedAt)
	require.NotEqual(t, requestedData.UpdatedAt, changedData.UpdatedAt)
	require.Equal(t, requestedData.AccessCount, changedData.AccessCount)

	// проверка удаления задачи
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest("DELETE", "/shorten/"+data.ShortCode, nil)
	req5.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w5, req5)
	
	require.Equal(t, http.StatusNoContent, w5.Code)
	_, err := app.UrlRepo.RepositoryGet(data.ShortCode, data.OwnerID)
	require.Error(t, err)
}

