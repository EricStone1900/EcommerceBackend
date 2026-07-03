package s3

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Storage implements port.Storage using S3-compatible object storage.
// Supports both AWS S3 and S3-compatible services (MinIO, Alibaba OSS, etc.).
type Storage struct {
	client     *http.Client
	endpoint   string
	bucket     string
	accessKey  string
	secretKey  string
	region     string
	logger     *zap.Logger
}

// Config holds the configuration for S3-compatible storage.
type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

// NewStorage creates a new S3-compatible storage backend.
func NewStorage(cfg Config, logger *zap.Logger) *Storage {
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	logger.Info("s3 storage initialized",
		zap.String("endpoint", endpoint),
		zap.String("bucket", cfg.Bucket),
		zap.String("region", cfg.Region),
	)
	return &Storage{
		client:    &http.Client{Timeout: 60 * time.Second},
		endpoint:  endpoint,
		bucket:    cfg.Bucket,
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
		region:    cfg.Region,
		logger:    logger,
	}
}

// Upload stores a file in S3-compatible storage and returns its URL.
// The URL is constructed as: {endpoint}/{bucket}/{filename}
func (s *Storage) Upload(ctx context.Context, data []byte, filename string) (string, error) {
	objectKey := path.Join(s.bucket, filename)
	reqURL := fmt.Sprintf("%s/%s", s.endpoint, objectKey)

	bodyReader := bytes.NewReader(data)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("s3: failed to create upload request: %w", err)
	}

	contentType := detectContentType(filename)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	// Sign the request with AWS Signature V4
	if err := s.signRequest(req, data); err != nil {
		return "", fmt.Errorf("s3: failed to sign request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("s3: upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("s3: upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	url := fmt.Sprintf("%s/%s", s.endpoint, objectKey)
	s.logger.Debug("s3 file uploaded",
		zap.String("url", url),
		zap.Int("size", len(data)),
	)
	return url, nil
}

// Delete removes a file from S3-compatible storage.
func (s *Storage) Delete(ctx context.Context, urlStr string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlStr, nil)
	if err != nil {
		return fmt.Errorf("s3: failed to create delete request: %w", err)
	}

	// Sign with empty body
	if err := s.signRequest(req, nil); err != nil {
		return fmt.Errorf("s3: failed to sign delete request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("s3: delete request failed: %w", err)
	}
	defer resp.Body.Close()

	// 404 means the object doesn't exist — treat as success (idempotent)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("s3: delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	s.logger.Debug("s3 file deleted", zap.String("url", urlStr))
	return nil
}

// detectContentType maps file extensions to MIME types.
func detectContentType(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt", ".md":
		return "text/plain"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

// --- AWS Signature V4 ---

// signRequest adds AWS Signature V4 authentication headers to the request.
// This works with AWS S3, MinIO, and any S3-compatible API.
func (s *Storage) signRequest(req *http.Request, body []byte) error {
	if body == nil {
		body = []byte{}
	}

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	// Payload hash
	payloadHash := sha256Hex(body)

	// Headers
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	// Canonical request
	canonicalURI := req.URL.Path
	canonicalQuery := req.URL.RawQuery

	// Sort headers
	var headerNames []string
	for name := range req.Header {
		headerNames = append(headerNames, strings.ToLower(name))
	}
	sort.Strings(headerNames)

	var signedHeaders []string
	for _, name := range headerNames {
		signedHeaders = append(signedHeaders, name)
	}

	canonicalHeaders := ""
	for _, name := range signedHeaders {
		canonicalHeaders += name + ":" + strings.TrimSpace(req.Header.Get(name)) + "\n"
	}

	signedHeadersStr := strings.Join(signedHeaders, ";")

	canonicalRequest := req.Method + "\n" +
		canonicalURI + "\n" +
		canonicalQuery + "\n" +
		canonicalHeaders + "\n" +
		signedHeadersStr + "\n" +
		payloadHash

	// String to sign
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := dateStamp + "/" + s.region + "/s3/aws4_request"
	stringToSign := algorithm + "\n" +
		amzDate + "\n" +
		credentialScope + "\n" +
		sha256Hex([]byte(canonicalRequest))

	// Signing key
	signingKey := s.getSignatureKey(s.secretKey, dateStamp, s.region, "s3")

	// Signature
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// Authorization header
	authHeader := algorithm + " " +
		"Credential=" + s.accessKey + "/" + credentialScope + ", " +
		"SignedHeaders=" + signedHeadersStr + ", " +
		"Signature=" + signature

	req.Header.Set("Authorization", authHeader)
	return nil
}

func (s *Storage) getSignatureKey(key, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+key), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(regionName))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
