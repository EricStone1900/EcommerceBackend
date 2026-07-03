package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	gormdb "github.com/EricStone1900/ecommerce-backend/internal/infrastructure/persistence/gorm"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/handler"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/auth"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/product"
	"github.com/EricStone1900/ecommerce-backend/pkg/jwt"
)

// ---------------------------------------------------------------------------
// Test-specific GORM models for auto-migration (mirrors the unexported models
// in internal/infrastructure/persistence/gorm so we can create the tables in
// an in-memory SQLite database).
// ---------------------------------------------------------------------------

type testUserModel struct {
	ID           uint           `gorm:"primaryKey"`
	Email        string         `gorm:"uniqueIndex;not null;size:255"`
	PasswordHash string         `gorm:"not null;size:255"`
	Role         string         `gorm:"not null;default:customer"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (testUserModel) TableName() string { return "users" }

type testProductModel struct {
	ID          uint           `gorm:"primaryKey"`
	Name        string         `gorm:"not null;size:255"`
	Description string         `gorm:"not null"`
	Price       float64        `gorm:"not null"`
	Stock       int            `gorm:"not null;default:0"`
	Status      string         `gorm:"not null;size:20;default:on_sale"`
	CreatedBy   uint           `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (testProductModel) TableName() string { return "products" }

type testFileModel struct {
	ID           uint           `gorm:"primaryKey"`
	OwnerID      uint           `gorm:"not null;index"`
	Type         string         `gorm:"not null;size:20"`
	OriginalName string         `gorm:"not null;size:255"`
	URL          string         `gorm:"not null;size:512"`
	Size         int64          `gorm:"not null;default:0"`
	Status       string         `gorm:"not null;size:20;default:pending"`
	CreatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (testFileModel) TableName() string { return "files" }

type testPushTokenModel struct {
	ID          uint           `gorm:"primaryKey"`
	UserID      uint           `gorm:"not null;uniqueIndex:idx_push_tokens_user_device"`
	DeviceToken string         `gorm:"not null;uniqueIndex:idx_push_tokens_user_device"`
	Platform    string         `gorm:"not null;size:20"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (testPushTokenModel) TableName() string { return "push_tokens" }

// ---------------------------------------------------------------------------
// In-memory TokenStore (replaces Redis for integration tests).
// ---------------------------------------------------------------------------

type inMemTokenStore struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

func newInMemTokenStore() *inMemTokenStore {
	return &inMemTokenStore{tokens: make(map[string]time.Time)}
}

func (s *inMemTokenStore) tokenKey(userID uint, tokenID string) string {
	return fmt.Sprintf("%d:%s", userID, tokenID)
}

func (s *inMemTokenStore) SaveRefreshToken(_ context.Context, userID uint, tokenID string, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[s.tokenKey(userID, tokenID)] = time.Now().Add(expiration)
	return nil
}

func (s *inMemTokenStore) ValidateRefreshToken(_ context.Context, userID uint, tokenID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.tokenKey(userID, tokenID)
	exp, ok := s.tokens[key]
	if !ok {
		return false, nil
	}
	if time.Now().After(exp) {
		delete(s.tokens, key)
		return false, nil
	}
	return true, nil
}

func (s *inMemTokenStore) DeleteRefreshToken(_ context.Context, userID uint, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, s.tokenKey(userID, tokenID))
	return nil
}

// ---------------------------------------------------------------------------
// Shared test dependencies.
// ---------------------------------------------------------------------------

const (
	testJWTSecret     = "test-secret-for-integration-tests"
	testAdminEmail    = "admin@test.com"
	testAdminPassword = "Admin@123"
)

// TestDependencies holds all shared infrastructure for integration tests.
type TestDependencies struct {
	DB          *gorm.DB
	UserRepo    *gormdb.UserRepository
	ProductRepo *gormdb.ProductRepository
	TokenStore  *inMemTokenStore
	JWTService  *jwt.JWTService
	Logger      *zap.Logger

	AuthUseCase    *auth.AuthUseCase
	ProductUseCase *product.ProductUseCase

	AuthHandler    *handler.AuthHandler
	ProductHandler *handler.ProductHandler
	AuthMiddleware gin.HandlerFunc
}

var deps *TestDependencies

// TestMain is the entry point for all integration tests in this package.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// 1. Connect to in-memory SQLite with silent logger to keep test output clean.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open in-memory SQLite: %v\n", err)
		os.Exit(1)
	}

	// 2. Auto-migrate all tables.
	if err := db.AutoMigrate(
		&testUserModel{},
		&testProductModel{},
		&testFileModel{},
		&testPushTokenModel{},
	); err != nil {
		fmt.Fprintf(os.Stderr, "failed to migrate tables: %v\n", err)
		os.Exit(1)
	}

	// 3. Seed an admin user so that admin-protected routes can be tested.
	if err := seedAdminUser(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to seed admin user: %v\n", err)
		os.Exit(1)
	}

	// 4. Initialize shared application dependencies.
	tokenStore := newInMemTokenStore()
	jwtSvc := jwt.NewJWTService(testJWTSecret, 15*time.Minute, 7*24*time.Hour)
	logger := zap.NewNop()
	userRepo := gormdb.NewUserRepository(db)
	productRepo := gormdb.NewProductRepository(db)

	authUC := auth.NewAuthUseCase(userRepo, tokenStore, jwtSvc, logger, 15*time.Minute, 7*24*time.Hour)
	productUC := product.NewProductUseCase(productRepo, logger)

	authH := handler.NewAuthHandler(authUC)
	productH := handler.NewProductHandler(productUC)
	authMW := middleware.AuthMiddleware(jwtSvc)

	deps = &TestDependencies{
		DB:             db,
		UserRepo:       userRepo,
		ProductRepo:    productRepo,
		TokenStore:     tokenStore,
		JWTService:     jwtSvc,
		Logger:         logger,
		AuthUseCase:    authUC,
		ProductUseCase: productUC,
		AuthHandler:    authH,
		ProductHandler: productH,
		AuthMiddleware: authMW,
	}

	// 5. Run all tests.
	code := m.Run()
	os.Exit(code)
}

// seedAdminUser inserts the test admin user directly into the database.
func seedAdminUser(db *gorm.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(testAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	now := time.Now()
	return db.Exec(
		`INSERT INTO users (email, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		testAdminEmail, string(hash), string(entity.RoleAdmin), now, now,
	).Error
}

// ---------------------------------------------------------------------------
// Test-helper functions.
// ---------------------------------------------------------------------------

// authTokens holds the tokens and user info returned by register/login.
type authTokens struct {
	AccessToken  string
	RefreshToken string
	UserID       uint
	UserEmail    string
	UserRole     entity.Role
}

// newAuthRouter creates a Gin engine with all auth routes registered (mirrors
// the route registration in cmd/server/main.go).
func newAuthRouter() *gin.Engine {
	r := gin.New()

	// Public routes
	r.POST("/api/v1/auth/register", deps.AuthHandler.Register)
	r.POST("/api/v1/auth/login", deps.AuthHandler.Login)

	// Protected routes (require valid JWT access token)
	authGroup := r.Group("/api/v1/auth")
	authGroup.Use(deps.AuthMiddleware)
	authGroup.GET("/me", deps.AuthHandler.Me)
	authGroup.POST("/refresh", deps.AuthHandler.Refresh)
	authGroup.POST("/logout", deps.AuthHandler.Logout)

	return r
}

// newProductRouter creates a Gin engine with product routes registered.
func newProductRouter() *gin.Engine {
	r := gin.New()

	// Public
	r.GET("/api/v1/products/:id", deps.ProductHandler.GetDetail)

	// Protected (any authenticated user)
	r.GET("/api/v1/products", deps.AuthMiddleware, deps.ProductHandler.List)

	// Protected (admin only)
	r.POST("/api/v1/products", deps.AuthMiddleware, middleware.RBACMiddleware(entity.RoleAdmin), deps.ProductHandler.Create)

	return r
}

// registerUser sends a register request through the provided router and returns
// the parsed auth tokens.
func registerUser(t testing.TB, router *gin.Engine, email, password string) authTokens {
	t.Helper()

	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "register should return 201")
	return parseAuthResponse(t, w.Body.Bytes())
}

// loginUser sends a login request through the provided router and returns the
// parsed auth tokens.
func loginUser(t testing.TB, router *gin.Engine, email, password string) authTokens {
	t.Helper()

	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "login should return 200")
	return parseAuthResponse(t, w.Body.Bytes())
}

// parseAuthResponse unmarshals the standard JSON response wrapper and extracts
// the auth tokens.
func parseAuthResponse(t testing.TB, raw []byte) authTokens {
	t.Helper()

	var wrapper struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data,omitempty"`
	}
	require.NoError(t, json.Unmarshal(raw, &wrapper))
	require.Equal(t, 0, wrapper.Code, "response code should be 0 (success)")

	var data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID    uint   `json:"id"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	require.NoError(t, json.Unmarshal(wrapper.Data, &data))

	return authTokens{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		UserID:       data.User.ID,
		UserEmail:    data.User.Email,
		UserRole:     entity.Role(data.User.Role),
	}
}

// httpDo is a convenience wrapper around httptest for making JSON/api requests.
func httpDo(t testing.TB, router *gin.Engine, method, path, body string, token string) *httptest.ResponseRecorder {
	t.Helper()

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
