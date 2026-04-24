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

// UploadURLResponse contains the presigned S3 upload URL.
type UploadURLResponse struct {
	FileID    string `json:"fileId"`
	UploadURL string `json:"uploadUrl"`
	Key       string `json:"key"`
	ExpiresAt string `json:"expiresAt"`
}

// StorageFile represents a stored file entry.
type StorageFile struct {
	FileID       string    `json:"fileId"`
	TenantID     string    `json:"tenantId"`
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
}

// ────────────────────────────────────────────────────────────────────────────
// Request types
// ────────────────────────────────────────────────────────────────────────────

// GenerateUploadURLRequest requests a presigned upload URL.
type GenerateUploadURLRequest struct {
	TenantID string `json:"tenantId"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
}

// ────────────────────────────────────────────────────────────────────────────
// StorageClient
// ────────────────────────────────────────────────────────────────────────────

// StorageClient provides access to the storage-service REST API.
type StorageClient struct {
	c *Client
}

// GenerateUploadURL returns a presigned S3 URL for uploading a file.
func (s *StorageClient) GenerateUploadURL(ctx context.Context, req GenerateUploadURLRequest) (*UploadURLResponse, error) {
	var out UploadURLResponse
	if err := s.c.do(ctx, "POST", "/v1/storage/upload-url", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFiles returns files belonging to a tenant.
func (s *StorageClient) ListFiles(ctx context.Context, tenantID string) ([]*StorageFile, error) {
	path := fmt.Sprintf("/v1/storage/files?tenantId=%s", url.QueryEscape(tenantID))
	var out []*StorageFile
	if err := s.c.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteFile removes a file from storage.
func (s *StorageClient) DeleteFile(ctx context.Context, fileID, tenantID string) error {
	path := fmt.Sprintf("/v1/storage/files/%s?tenantId=%s",
		url.PathEscape(fileID), url.QueryEscape(tenantID))
	return s.c.do(ctx, "DELETE", path, nil, nil)
}

// GetDownloadURL returns a presigned download URL or CDN URL for a file.
func (s *StorageClient) GetDownloadURL(ctx context.Context, fileID, tenantID string) (string, error) {
	path := fmt.Sprintf("/v1/storage/files/%s/url?tenantId=%s",
		url.PathEscape(fileID), url.QueryEscape(tenantID))
	var out struct {
		URL string `json:"url"`
	}
	if err := s.c.do(ctx, "GET", path, nil, &out); err != nil {
		return "", err
	}
	return out.URL, nil
}
