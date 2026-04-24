package service

import (
	"context"
	"testing"
	"time"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/achemerzaev/url-shortening-service/pkg/logger"
	"github.com/stretchr/testify/require"
)

type mockURLRepository struct {
	db     map[string]models.UrlInfo
	LastId int
}

type mockRedisURLRepository struct {
	db map[string]models.UrlInfo
}

func (r *mockRedisURLRepository) SaveUrl(ctx context.Context, data models.UrlInfo) error {
	newURL := models.UrlInfo{
		Id:          data.Id,
		Url:         data.Url,
		ShortCode:   data.ShortCode,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
		AccessCount: data.AccessCount,
		OwnerID:     data.OwnerID,
	}

	r.db[newURL.ShortCode] = newURL
	return nil
}

func (r *mockRedisURLRepository) GetUrl(ctx context.Context, shortCode string, ownerID int) (string, error) {
	if v, ok := r.db[shortCode]; !ok {
		return "", appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return "", appErr.ErrForbidden
	} else {
		return v.Url, nil
	}
}

func (r *mockRedisURLRepository) GetUrlStats(ctx context.Context, shortCode string, ownerID int) (models.UrlInfo, error) {
	if v, ok := r.db[shortCode]; !ok {
		return models.UrlInfo{}, appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return models.UrlInfo{}, appErr.ErrForbidden
	} else {
		return v, nil
	}
}

func (r *mockRedisURLRepository) UpdateUrl(ctx context.Context, requestedCode, newLongUrl string, updatedAt time.Time, ownerID int) (models.UrlInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.UrlInfo{}, appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return models.UrlInfo{}, appErr.ErrForbidden
	} else {
		v.Url = newLongUrl
		v.UpdatedAt = updatedAt
		return v, nil
	}

}

func (r *mockRedisURLRepository) DeleteUrl(ctx context.Context, shortCode string, ownerID int) error {
	if v, ok := r.db[shortCode]; !ok {
		return appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return appErr.ErrForbidden
	} else {
		delete(r.db, shortCode)
		return nil
	}

}

func setupurlsvc() *URLService {
	repo := &mockURLRepository{
		db: make(map[string]models.UrlInfo),
	}

	redisrepo := &mockRedisURLRepository{
		db: make(map[string]models.UrlInfo),
	}

	logger, _ := logger.New("debug")
	return NewUrlService(repo, redisrepo, logger)
}

func (r *mockURLRepository) RepositoryPost(ctx context.Context, data models.UrlInfo) (models.UrlInfo, error) {
	r.LastId += 1
	newURL := models.UrlInfo{
		Id:          r.LastId,
		Url:         data.Url,
		ShortCode:   data.ShortCode,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
		AccessCount: data.AccessCount,
		OwnerID:     data.OwnerID,
	}
	r.db[newURL.ShortCode] = newURL
	return newURL, nil
}

func (r *mockURLRepository) RepositoryGet(ctx context.Context, requestedCode string) (models.UrlInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.UrlInfo{}, appErr.ErrNotFound
	} else {
		return v, nil
	}
}

func (r *mockURLRepository) RepositoryUpdate(ctx context.Context, requestedCode string, longurl string, updatedAt time.Time, ownerID int) (models.UrlInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.UrlInfo{}, appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return models.UrlInfo{}, appErr.ErrForbidden
	} else {
		v.Url = longurl
		v.UpdatedAt = updatedAt
		return v, nil
	}

}

func (r *mockURLRepository) RepositoryDelete(ctx context.Context, requestedCode string, ownerID int) error {
	if v, ok := r.db[requestedCode]; !ok {
		return appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return appErr.ErrForbidden
	} else {
		delete(r.db, requestedCode)
		return nil
	}

}

func (r *mockURLRepository) RepositoryGetStats(ctx context.Context, requestedCode string) (models.UrlInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.UrlInfo{}, appErr.ErrNotFound
	} else {
		return v, nil
	}

}

func TestServcePost_Success(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, err := svc.ServicePost(context.Background(), insert)

	require.NoError(t, err)
	require.Equal(t, insert.Id, info.Id)
	require.Equal(t, insert.Url, info.Url)
	require.Equal(t, info.CreatedAt, info.UpdatedAt)
	require.Equal(t, insert.AccessCount, info.AccessCount)
	require.Equal(t, insert.OwnerID, info.OwnerID)
}

func TestServcePost_AddPrefix(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "url.com",
		ShortCode:   "shortc",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, err := svc.ServicePost(context.Background(), insert)

	require.NoError(t, err)
	require.Equal(t, info.Url, "https://url.com")
}

func TestServiceGet_Success(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	url, err := svc.ServiceGet(context.Background(), info.ShortCode, insert.OwnerID)

	require.NoError(t, err)
	require.Equal(t, url, insert.Url)

}

func TestServiceGet_Forbidden(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	url, err := svc.ServiceGet(context.Background(), info.ShortCode, 2)

	require.ErrorIs(t, err, appErr.ErrForbidden)
	require.Equal(t, url, "")
}

func TestServiceGet_NotFound(t *testing.T) {
	svc := setupurlsvc()

	url, err := svc.ServiceGet(context.Background(), "shortc", 1)

	require.ErrorIs(t, err, appErr.ErrNotFound)
	require.Equal(t, url, "")

}

func TestServicePut_Success(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		ShortCode:   "shortc",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	info1, err := svc.ServicePut(context.Background(), info.ShortCode, "http://longurl.com", 1)

	require.NoError(t, err)
	require.Equal(t, info1.Url, "http://longurl.com")
}

func TestServicePut_Forbidden(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	info, err := svc.ServicePut(context.Background(), info.ShortCode, "http://longurl.com", 2)

	require.ErrorIs(t, err, appErr.ErrForbidden)
	require.Equal(t, info, models.UrlInfo{})
}

func TestServicePut_NotFound(t *testing.T) {
	svc := setupurlsvc()
	info, err := svc.ServicePut(context.Background(), "shortc", "http://longurl.com", 1)

	require.ErrorIs(t, err, appErr.ErrNotFound)
	require.Equal(t, info, models.UrlInfo{})
}

func TestServiceDelete_Success(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	err := svc.ServiceDelete(context.Background(), info.ShortCode, 1)

	require.NoError(t, err)
}

func TestServiceDelete_Forbidden(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		ShortCode:   "shortc",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	err := svc.ServiceDelete(context.Background(), info.ShortCode, 2)

	require.ErrorIs(t, err, appErr.ErrForbidden)
}

func TestServiceDelete_NotFound(t *testing.T) {
	svc := setupurlsvc()
	err := svc.ServiceDelete(context.Background(), "shortc", 1)

	require.ErrorIs(t, err, appErr.ErrNotFound)
}

func TestServiceGetStats_Success(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	_, err := svc.ServiceGetStats(context.Background(), info.ShortCode, 1)

	require.NoError(t, err)
}

func TestServiceGetStats_SuccessUpdate(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		ShortCode:   "shortc",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	_, _ = svc.ServiceGet(context.Background(), info.ShortCode, 1)

	info1, err := svc.ServiceGetStats(context.Background(), info.ShortCode, 1)

	require.NoError(t, err)
	require.Equal(t, info1.AccessCount, 0)
}

func TestServiceGetStats_Forbidden(t *testing.T) {
	svc := setupurlsvc()

	insert := models.UrlInfo{
		Id:          1,
		Url:         "http://url.com",
		ShortCode:   "shortc",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	_, err := svc.ServiceGetStats(context.Background(), info.ShortCode, 2)

	require.ErrorIs(t, err, appErr.ErrForbidden)
}

func TestServiceGetStats_NotFound(t *testing.T) {
	svc := setupurlsvc()
	_, err := svc.ServiceGetStats(context.Background(), "shortc", 1)

	require.ErrorIs(t, err, appErr.ErrNotFound)
}
