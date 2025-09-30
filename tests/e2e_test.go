package e2e

import (
	"github.com/boretsotets/url-shortening-service/database"
	"github.com/boretsotets/url-shortening-service/internal/authorization"
	"github.com/boretsotets/url-shortening-service/internal/handler"
	"github.com/boretsotets/url-shortening-service/internal/models"
	"github.com/boretsotets/url-shortening-service/internal/redisrepo"
	"github.com/boretsotets/url-shortening-service/internal/repository"
	"github.com/boretsotets/url-shortening-service/internal/service"
	"github.com/boretsotets/url-shortening-service/internal/middleware"


	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestApp struct {
	Router   *gin.Engine
	UserRepo *repository.UserRepository
	UrlRepo  *repository.UrlRepository
}

func setupTestApp(t *testing.T) *TestApp {
	t.Helper()

	ctx := context.Background()

	dsn := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
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

	router := gin.New()
	router.Use(middleware.RequestIdMiddleware())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.JSONValidationMiddleware())

	router.POST("/register", userHandler.HandlerRegister)
	router.POST("/login", userHandler.HandlerLogin)
	router.POST("/refresh", userHandler.HandlerRefresh)

	private := router.Group("/")
	private.Use(middleware.AuthorizationMiddleware(logger))
	{
		private.POST("/shorten", urlHandler.HandlerPost)
		private.GET("/shorten/:shortcode", urlHandler.HandlerGet)
		private.PUT("/shorten/:shortcode", urlHandler.HandlerPut)
		private.DELETE("/shorten/:shortcode", urlHandler.HandlerDelete)
		private.GET("/shorten/:shortcode/stats", urlHandler.HandlerGetStats)
	}

	return &TestApp{
		Router:   router,
		UserRepo: userRepo,
		UrlRepo:  urlRepo,
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
	_ = json.NewDecoder(w.Body).Decode(&tokens)
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
	refreshtokenrequest := `{"refresh_token": "` + tokens.RefreshToken + `"}`
	req2 := httptest.NewRequest("POST", "/refresh", strings.NewReader(refreshtokenrequest))
	app.Router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusCreated, w2.Code)

	// проверка, что не валидный рефреш токен не работает
	w2 = httptest.NewRecorder()
	req2 = httptest.NewRequest("POST", "/refresh", strings.NewReader(refreshtokenrequest))
	app.Router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusUnauthorized, w2.Code)

	// проверка логина
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"bb", "password":"c"}`))
	req3.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusOK, w3.Code)

	var tokens1 models.Tokens
	_ = json.NewDecoder(w3.Body).Decode(&tokens1)
	require.NotEmpty(t, tokens1.AccessToken)
	require.NotEmpty(t, tokens1.RefreshToken)

}

func TestEmailDuplicate(t *testing.T) {
	app := setupTestApp(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"a", "email":"bb", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestInvalidToken(t *testing.T) {
	app := setupTestApp(t)

	AccessToken := "unvalid.access.token"
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "mail.ru"}`))
	req.Header.Set("Authorization", AccessToken)
	app.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHasNoAccess(t *testing.T) {
	app := setupTestApp(t)

	// создание первого юзера
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"user1", "email":"user1", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)

	// создание короткой ссылки первого юзера
	var tokens models.Tokens
	_ = json.NewDecoder(w.Body).Decode(&tokens)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "mail.ru"}`))
	req1.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusOK, w1.Code)
	var data models.UrlInfo
	_ = json.NewDecoder(w1.Body).Decode(&data)


	// создание второго юзера
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"user2", "email":"user2", "password":"c"}`))
	req2.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w2, req2)

	var tokens2 models.Tokens
	_ = json.NewDecoder(w2.Body).Decode(&tokens2)

	// попытка доступа второго юзера к ссылке первого юзера
		// проверка редиректа
		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest("GET", "/shorten/"+data.ShortCode, nil)
		req3.Header.Set("Authorization", tokens2.AccessToken)
		app.Router.ServeHTTP(w3, req3)
		require.Equal(t, http.StatusForbidden, w3.Code)

		// проверка получения статистики
		w4 := httptest.NewRecorder()
		req4 := httptest.NewRequest("GET", "/shorten/"+data.ShortCode+"/stats", nil)
		req4.Header.Set("Authorization", tokens2.AccessToken)
		app.Router.ServeHTTP(w4, req4)
		require.Equal(t, http.StatusForbidden, w4.Code)

		// проверка изменения задачи
		w5 := httptest.NewRecorder()
		req5 := httptest.NewRequest("PUT", "/shorten/"+data.ShortCode, strings.NewReader(`{"url": "go.dev"}`))
		req5.Header.Set("Authorization", tokens2.AccessToken)
		app.Router.ServeHTTP(w5, req5)
		require.Equal(t, http.StatusForbidden, w5.Code)

		// проверка удаления задачи
		w6 := httptest.NewRecorder()
		req6 := httptest.NewRequest("DELETE", "/shorten/"+data.ShortCode, nil)
		req6.Header.Set("Authorization", tokens2.AccessToken)
		app.Router.ServeHTTP(w6, req6)
		require.Equal(t, http.StatusForbidden, w6.Code)
}

func TestCrudOperations(t *testing.T) {
	app := setupTestApp(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"a", "email":"bbb", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)

	var tokens models.Tokens
	_ = json.NewDecoder(w.Body).Decode(&tokens)

	// проверка создания короткой ссылки
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "mail.ru"}`))
	req1.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusOK, w1.Code)
	var data models.UrlInfo
	_ = json.NewDecoder(w1.Body).Decode(&data)

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
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("error closing body: %s", err)
		}
	}()

	loc, _ := res.Location()
	require.Equal(t, "https://mail.ru", loc.String())

	// проверка получения статистики
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/shorten/"+data.ShortCode+"/stats", nil)
	req3.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w3, req3)

	var requestedData models.UrlInfo
	_ = json.NewDecoder(w3.Body).Decode(&requestedData)
	fmt.Print("date here: ", requestedData.UpdatedAt)
	fmt.Print("################################################")

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
	_ = json.NewDecoder(w4.Body).Decode(&changedData)

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
	_, err := app.UrlRepo.RepositoryGet(data.ShortCode)
	require.Error(t, err)
}

func TestNotInDatabase(t *testing.T) {
	app := setupTestApp(t)

	// создание юзера
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"user3", "email":"user3", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)

	var tokens models.Tokens
	_ = json.NewDecoder(w.Body).Decode(&tokens)

	// проверка ошибки при отсуствтующем коротком коде
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/shorten/"+"shortcode", nil)
	req1.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusNotFound, w1.Code)

	// создание, удаление и проверка отсутствия
		// создание короткой ссылки
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "google.com"}`))
		req2.Header.Set("Authorization", tokens.AccessToken)
		app.Router.ServeHTTP(w2, req2)
		require.Equal(t, http.StatusOK, w2.Code)
		var data models.UrlInfo
		_ = json.NewDecoder(w2.Body).Decode(&data)
		shortCode := data.ShortCode

		// удаление ссылки
		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest("DELETE", "/shorten/"+shortCode, nil)
		req3.Header.Set("Authorization", tokens.AccessToken)
		app.Router.ServeHTTP(w3, req3)
		require.Equal(t, http.StatusNoContent, w3.Code)

		// попытка доступа к удаленной ссылке
		w4 := httptest.NewRecorder()
		req4 := httptest.NewRequest("GET", "/shorten/"+shortCode, nil)
		req4.Header.Set("Authorization", tokens.AccessToken)
		app.Router.ServeHTTP(w4, req4)
		require.Equal(t, http.StatusNotFound, w4.Code)

		// проверка получения статистики
		w5 := httptest.NewRecorder()
		req5 := httptest.NewRequest("GET", "/shorten/"+shortCode+"/stats", nil)
		req5.Header.Set("Authorization", tokens.AccessToken)
		app.Router.ServeHTTP(w5, req5)
		require.Equal(t, http.StatusNotFound, w5.Code)
}

func TestInvalidJSON(t *testing.T) {
	app := setupTestApp(t)

	// создание юзера
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"nam":"user4", "email":"user4", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"user4", "emai":"user4", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"user4", "email":"user4", "passwor":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"user4", "email":"user4", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var tokens models.Tokens
	_ = json.NewDecoder(w.Body).Decode(&tokens)

	// логин
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/login", strings.NewReader(`{"emai":"user4", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"user4", "passwor":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"user4", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// рефреш
	var tokens1 models.Tokens
	_ = json.NewDecoder(w.Body).Decode(&tokens1)
	require.NotEmpty(t, tokens1.AccessToken)
	require.NotEmpty(t, tokens1.RefreshToken)

	w = httptest.NewRecorder()
	refreshtokenrequest := `{"refreshtoken": "` + tokens1.RefreshToken + `"}`
	req = httptest.NewRequest("POST", "/refresh", strings.NewReader(refreshtokenrequest))
	app.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// создание короткой ссылки
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"urll": "mail.ru"}`))
	req1.Header.Set("Authorization", tokens1.AccessToken)
	app.Router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusBadRequest, w1.Code)

	w1 = httptest.NewRecorder()
	req1 = httptest.NewRequest("POST", "/shorten", strings.NewReader(`{}`))
	req1.Header.Set("Authorization", tokens1.AccessToken)
	app.Router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusBadRequest, w1.Code)

	w1 = httptest.NewRecorder()
	req1 = httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "mail.ru"}`))
	req1.Header.Set("Authorization", tokens1.AccessToken)
	app.Router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code)

	var data models.UrlInfo
	_ = json.NewDecoder(w1.Body).Decode(&data)

	// изменение ссылки
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("PUT", "/shorten/"+data.ShortCode, strings.NewReader(`{"ur": "go.dev"}`))
	req2.Header.Set("Authorization", tokens1.AccessToken)
	app.Router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusBadRequest, w2.Code)

	w2 = httptest.NewRecorder()
	req2 = httptest.NewRequest("PUT", "/shorten/"+data.ShortCode, strings.NewReader(`{}`))
	req2.Header.Set("Authorization", tokens1.AccessToken)
	app.Router.ServeHTTP(w2, req2)	
	require.Equal(t, http.StatusBadRequest, w2.Code)

}


func TestRedis(t *testing.T) {
	app := setupTestApp(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"a", "email":"redis_test", "password":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	app.Router.ServeHTTP(w, req)

	var tokens models.Tokens
	_ = json.NewDecoder(w.Body).Decode(&tokens)

	// создание короткой ссылки
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "mail.ru"}`))
	req1.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w1, req1)
	var data models.UrlInfo
	_ = json.NewDecoder(w1.Body).Decode(&data)

	// редирект - после этого обращение к редису
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/shorten/"+data.ShortCode, nil)
	req2.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusFound, w2.Code)

	// редирект - обращение к редису
	w2 = httptest.NewRecorder()
	req2 = httptest.NewRequest("GET", "/shorten/"+data.ShortCode, nil)
	req2.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusFound, w2.Code)

	// статистика из редиса
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/shorten/"+data.ShortCode+"/stats", nil)
	req3.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusOK, w3.Code)

	// проверка изменения задачи в редисе
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("PUT", "/shorten/"+data.ShortCode, strings.NewReader(`{"url": "go.dev"}`))
	req4.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w4, req4)
	require.Equal(t, http.StatusOK, w4.Code)

	// проверка удаления задачи
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest("DELETE", "/shorten/"+data.ShortCode, nil)
	req5.Header.Set("Authorization", tokens.AccessToken)
	app.Router.ServeHTTP(w5, req5)
	require.Equal(t, http.StatusNoContent, w5.Code)

	// проверка ошибки forbidden
		// создание короткой ссылки
		w1 = httptest.NewRecorder()
		req1 = httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "mail.ru"}`))
		req1.Header.Set("Authorization", tokens.AccessToken)
		app.Router.ServeHTTP(w1, req1)
		var data1 models.UrlInfo
		_ = json.NewDecoder(w1.Body).Decode(&data1)

		// гет статс для сохранения в редисе
		w2 = httptest.NewRecorder()
		req2 = httptest.NewRequest("GET", "/shorten/"+data1.ShortCode+"/stats", nil)
		req2.Header.Set("Authorization", tokens.AccessToken)
		app.Router.ServeHTTP(w2, req2)
		require.Equal(t, http.StatusOK, w2.Code)

		// создание нового юзера
		w3 = httptest.NewRecorder()
		req3 = httptest.NewRequest("POST", "/register", strings.NewReader(`{"name":"a", "email":"redis_test1", "password":"c"}`))
		req3.Header.Set("Content-Type", "application/json")
		app.Router.ServeHTTP(w3, req3)
		var tokens1 models.Tokens
		_ = json.NewDecoder(w3.Body).Decode(&tokens1)
	
		// редирект
		w3 = httptest.NewRecorder()
		req3 = httptest.NewRequest("GET", "/shorten/"+data1.ShortCode, nil)
		req3.Header.Set("Authorization", tokens1.AccessToken)
		app.Router.ServeHTTP(w3, req3)
		require.Equal(t, http.StatusForbidden, w3.Code)

		// статистика
		w3 = httptest.NewRecorder()
		req3 = httptest.NewRequest("GET", "/shorten/"+data1.ShortCode+"/stats", nil)
		req3.Header.Set("Authorization", tokens1.AccessToken)
		app.Router.ServeHTTP(w3, req3)
		require.Equal(t, http.StatusForbidden, w3.Code)

		// изменение задачи
		w3 = httptest.NewRecorder()
		req3 = httptest.NewRequest("PUT", "/shorten/"+data1.ShortCode, strings.NewReader(`{"url": "go.dev"}`))
		req3.Header.Set("Authorization", tokens1.AccessToken)
		app.Router.ServeHTTP(w3, req3)
		require.Equal(t, http.StatusForbidden, w3.Code)

		// удаление задачи
		w3 = httptest.NewRecorder()
		req3 = httptest.NewRequest("DELETE", "/shorten/"+data1.ShortCode, nil)
		req3.Header.Set("Authorization", tokens1.AccessToken)
		app.Router.ServeHTTP(w3, req3)
		require.Equal(t, http.StatusForbidden, w3.Code)
}
