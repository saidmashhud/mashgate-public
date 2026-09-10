package mashgate

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Domain types
// ────────────────────────────────────────────────────────────────────────────

// UploadURLResponse contains the presigned upload URL.
//
// Поля — из контракта (contracts/proto/v1/storage.proto, GetUploadUrlResponse:
// upload_url, object_id). Прежняя версия объявляла fileId/key/expiresAt,
// которых сервис не возвращает: они приходили пустыми, а вызывающий считал,
// что получил идентификатор объекта.
type UploadURLResponse struct {
	UploadURL string `json:"uploadUrl"`
	ObjectID  string `json:"objectId"`
}

// StorageObject — объект в хранилище.
//
// Поля из контракта (StorageObject: id, bucket_id, key, size_bytes,
// content_type, created_at). Прежний StorageFile объявлял fileId, size и
// lastModified — таких полей сервис не возвращает, все три приходили нулевыми.
type StorageObject struct {
	ID          string    `json:"id"`
	BucketID    string    `json:"bucketId"`
	Key         string    `json:"key"`
	SizeBytes   int64     `json:"sizeBytes"`
	ContentType string    `json:"contentType"`
	CreatedAt   time.Time `json:"createdAt"`
}

// StorageFile — прежнее имя типа. Оставлено псевдонимом, чтобы существующий
// код собирался; поля у него теперь правильные.
type StorageFile = StorageObject

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// GenerateUploadURLRequest requests a presigned upload URL.
//
// Поля — из контракта (GetUploadUrlRequest: tenant_id, bucket_id, key,
// content_type), и они же в сгенерированном слое mashgatev1.
//
// Прежняя версия слала filename и mimeType — таких полей в договоре нет.
// Оба доезжали ПУСТЫМИ, и это не было безобидно: ключ объекта собирается как
// «тенант/идентификатор/имя», без имени он оканчивался косой чертой, подпись
// считалась по пустому типу содержимого — и любая закачка отвергалась с
// SignatureDoesNotMatch, сколько бы раз её ни повторяли. Grid уже нашёл это
// у себя и починил в обход SDK; здесь тот же дефект жил в самом SDK.
type GenerateUploadURLRequest struct {
	TenantID string `json:"tenantId"`
	// Key — полный ключ объекта в хранилище, а не имя файла.
	Key string `json:"key"`
	// ContentType нужен для подписи: она считается вместе с ним, и пустое
	// значение делает подпись недействительной.
	ContentType string `json:"contentType"`
	// BucketID необязателен: сервис берёт корзину тенанта по умолчанию.
	BucketID string `json:"bucketId,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// StorageClient
// ────────────────────────────────────────────────────────────────────────────

// StorageClient provides access to the storage-service REST API.
type StorageClient struct {
	c *Client
}

// GenerateUploadURL возвращает подписанный адрес для загрузки объекта.
func (s *StorageClient) GenerateUploadURL(ctx context.Context, req GenerateUploadURLRequest) (*UploadURLResponse, error) {
	var out UploadURLResponse
	if err := s.c.do(ctx, "POST", "/v1/storage/upload-url", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListObjects возвращает объекты корзины.
//
// Путь из контракта: GET /v1/storage/buckets/{bucket_id}/objects. Прежний
// ListFiles бил в /v1/storage/files, которого у сервиса нет вовсе — вызов
// возвращал ошибку маршрута при любом тенанте.
func (s *StorageClient) ListObjects(ctx context.Context, tenantID, bucketID string) ([]*StorageObject, error) {
	path := fmt.Sprintf("/v1/storage/buckets/%s/objects?tenantId=%s",
		url.PathEscape(bucketID), url.QueryEscape(tenantID))
	var out struct {
		Objects    []*StorageObject `json:"objects"`
		TotalCount int              `json:"totalCount"`
	}
	if err := s.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Objects, nil
}

// ListFiles — прежнее имя. Требует корзину: без неё контракт не отдаёт список.
//
// Deprecated: используйте ListObjects.
func (s *StorageClient) ListFiles(ctx context.Context, tenantID, bucketID string) ([]*StorageObject, error) {
	return s.ListObjects(ctx, tenantID, bucketID)
}

// GetObject возвращает один объект.
func (s *StorageClient) GetObject(ctx context.Context, objectID, tenantID string) (*StorageObject, error) {
	path := fmt.Sprintf("/v1/storage/objects/%s?tenantId=%s",
		url.PathEscape(objectID), url.QueryEscape(tenantID))
	var out StorageObject
	if err := s.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteObject удаляет объект.
func (s *StorageClient) DeleteObject(ctx context.Context, objectID, tenantID string) error {
	path := fmt.Sprintf("/v1/storage/objects/%s?tenantId=%s",
		url.PathEscape(objectID), url.QueryEscape(tenantID))
	return s.c.do(ctx, "DELETE", path, nil, nil)
}

// DeleteFile — прежнее имя.
//
// Deprecated: используйте DeleteObject.
func (s *StorageClient) DeleteFile(ctx context.Context, objectID, tenantID string) error {
	return s.DeleteObject(ctx, objectID, tenantID)
}

// GetDownloadURL возвращает подписанный адрес для скачивания.
//
// Ответ по контракту — download_url и expires_in_seconds; прежняя версия
// читала поле url, которого в ответе нет, и всегда возвращала пустую строку
// без ошибки. Тот же вид отказа, что мы ловим весь день: успех без результата.
func (s *StorageClient) GetDownloadURL(ctx context.Context, objectID, tenantID string) (string, error) {
	path := fmt.Sprintf("/v1/storage/objects/%s/download?tenantId=%s",
		url.PathEscape(objectID), url.QueryEscape(tenantID))
	var out struct {
		DownloadURL      string `json:"downloadUrl"`
		ExpiresInSeconds int64  `json:"expiresInSeconds"`
	}
	if err := s.c.do(ctx, "GET", path, nil, &out); err != nil {
		return "", err
	}
	if out.DownloadURL == "" {
		return "", fmt.Errorf("storage: сервис не вернул адрес скачивания для объекта %s", objectID)
	}
	return out.DownloadURL, nil
}
