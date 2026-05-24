package service

import (
	"context"
	"testing"
	"time"

	"github.com/achemerzaev/url-shortening-service/internal/models"
	appErr "github.com/achemerzaev/url-shortening-service/pkg/errors"
	"github.com/stretchr/testify/require"
)

type mockURLRepository struct {
	db     map[string]models.URLInfo
	LastId int
}

type mockRedisURLRepository struct {
	db map[string]models.URLInfo
}

func (r *mockRedisURLRepository) SaveUrl(ctx context.Context, data models.URLInfo) error {
	newURL := models.URLInfo{
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

func (r *mockRedisURLRepository) IncrementCounter(ctx context.Context, shortCode string) error {
	v := r.db[shortCode]
	v.AccessCount += 1
	r.db[shortCode] = v
	return nil
}

func (r *mockRedisURLRepository) GetUrlStats(ctx context.Context, shortCode string, ownerID int) (models.URLInfo, error) {
	if v, ok := r.db[shortCode]; !ok {
		return models.URLInfo{}, appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return models.URLInfo{}, appErr.ErrForbidden
	} else {
		return v, nil
	}
}

func (r *mockRedisURLRepository) UpdateUrl(ctx context.Context, requestedCode, newlongURL string, updatedAt time.Time, ownerID int) (models.URLInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.URLInfo{}, appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return models.URLInfo{}, appErr.ErrForbidden
	} else {
		v.Url = newlongURL
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
		db: make(map[string]models.URLInfo),
	}

	redisrepo := &mockRedisURLRepository{
		db: make(map[string]models.URLInfo),
	}

	return NewUrlService(repo, redisrepo)
}

func (r *mockURLRepository) RepositoryPost(ctx context.Context, data models.URLInfo) (models.URLInfo, error) {
	r.LastId += 1
	newURL := models.URLInfo{
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

func (r *mockURLRepository) RepositoryGet(ctx context.Context, requestedCode string) (models.URLInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.URLInfo{}, appErr.ErrNotFound
	} else {
		return v, nil
	}
}

func (r *mockURLRepository) RepositoryUpdate(ctx context.Context, requestedCode string, longURL string, updatedAt time.Time, ownerID int) (models.URLInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.URLInfo{}, appErr.ErrNotFound
	} else if v.OwnerID != ownerID {
		return models.URLInfo{}, appErr.ErrForbidden
	} else {
		v.Url = longURL
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

func (r *mockURLRepository) RepositoryGetStats(ctx context.Context, requestedCode string) (models.URLInfo, error) {
	if v, ok := r.db[requestedCode]; !ok {
		return models.URLInfo{}, appErr.ErrNotFound
	} else {
		return v, nil
	}

}

func TestServcePost_Success(t *testing.T) {
	svc := setupurlsvc()

	insert := models.URLInfo{
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

	insert := models.URLInfo{
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

	insert := models.URLInfo{
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

	insert := models.URLInfo{
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

	insert := models.URLInfo{
		Id:          1,
		Url:         "http://url.com",
		ShortCode:   "shortc",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	info1, err := svc.ServicePut(context.Background(), info.ShortCode, "http://longURL.com", 1)

	require.NoError(t, err)
	require.Equal(t, info1.Url, "http://longURL.com")
}

func TestServicePut_Forbidden(t *testing.T) {
	svc := setupurlsvc()

	insert := models.URLInfo{
		Id:          1,
		Url:         "http://url.com",
		AccessCount: 0,
		OwnerID:     1,
	}

	info, _ := svc.ServicePost(context.Background(), insert)
	info, err := svc.ServicePut(context.Background(), info.ShortCode, "http://longURL.com", 2)

	require.ErrorIs(t, err, appErr.ErrForbidden)
	require.Equal(t, info, models.URLInfo{})
}

func TestServicePut_NotFound(t *testing.T) {
	svc := setupurlsvc()
	info, err := svc.ServicePut(context.Background(), "shortc", "http://longURL.com", 1)

	require.ErrorIs(t, err, appErr.ErrNotFound)
	require.Equal(t, info, models.URLInfo{})
}

func TestServiceDelete_Success(t *testing.T) {
	svc := setupurlsvc()

	insert := models.URLInfo{
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

	insert := models.URLInfo{
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

	insert := models.URLInfo{
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

	insert := models.URLInfo{
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
	require.Equal(t, info1.AccessCount, 1)
}

func TestServiceGetStats_Forbidden(t *testing.T) {
	svc := setupurlsvc()

	insert := models.URLInfo{
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
