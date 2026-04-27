package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
	"github.com/moontv/server/pkg/apikey"
)

func setupTestRouter(t *testing.T, secret []byte, prefix string) (*gin.Engine, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	tmp := t.TempDir()
	database.Init(tmp + "/test.db")

	r := gin.New()
	r.GET("/protected", APIKeyAuth(secret, prefix), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	cleanup := func() {
		db, _ := database.DB.DB()
		db.Close()
	}
	return r, cleanup
}

func TestAPIKeyAuthRejectsOldRotatedKey(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	prefix := "mtv_"

	r, cleanup := setupTestRouter(t, secret, prefix)
	defer cleanup()

	user := model.User{
		Username:     "alice",
		PasswordHash: "hash",
		Role:         "user",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	oldPlain, oldCipher, err := apikey.Generate(secret, user.ID, prefix)
	if err != nil {
		t.Fatalf("generate old key: %v", err)
	}

	user.APIKeyCipher = oldCipher
	if err := database.DB.Save(&user).Error; err != nil {
		t.Fatalf("save old cipher: %v", err)
	}

	// Verify old key works before rotation
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", oldPlain)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("old key should work before rotation: got %d, body=%s", w.Code, w.Body.String())
	}

	// Rotate: generate new key, save new cipher
	_, newCipher, err := apikey.Generate(secret, user.ID, prefix)
	if err != nil {
		t.Fatalf("generate new key: %v", err)
	}
	user.APIKeyCipher = newCipher
	if err := database.DB.Save(&user).Error; err != nil {
		t.Fatalf("save new cipher: %v", err)
	}

	// Old key must now be rejected
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", oldPlain)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("old key must be rejected after rotation: got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestAPIKeyAuthRejectsRevokedKey(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	prefix := "mtv_"

	r, cleanup := setupTestRouter(t, secret, prefix)
	defer cleanup()

	user := model.User{
		Username:     "bob",
		PasswordHash: "hash",
		Role:         "user",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	plain, cipher, err := apikey.Generate(secret, user.ID, prefix)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	user.APIKeyCipher = cipher
	if err := database.DB.Save(&user).Error; err != nil {
		t.Fatalf("save cipher: %v", err)
	}

	// Revoke
	user.APIKeyCipher = ""
	if err := database.DB.Save(&user).Error; err != nil {
		t.Fatalf("revoke: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", plain)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("revoked key must be rejected: got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestAPIKeyAuthRejectsBannedUser(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	prefix := "mtv_"

	r, cleanup := setupTestRouter(t, secret, prefix)
	defer cleanup()

	user := model.User{
		Username:     "charlie",
		PasswordHash: "hash",
		Role:         "user",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	plain, cipher, err := apikey.Generate(secret, user.ID, prefix)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	user.APIKeyCipher = cipher
	user.Banned = true
	if err := database.DB.Save(&user).Error; err != nil {
		t.Fatalf("save: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", plain)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("banned user must get 403: got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestAPIKeyAuthRejectsMissingKey(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	prefix := "mtv_"

	r, cleanup := setupTestRouter(t, secret, prefix)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("missing key must get 401: got %d", w.Code)
	}
}

func TestAPIKeyAuthRejectsInvalidKey(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	prefix := "mtv_"

	r, cleanup := setupTestRouter(t, secret, prefix)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", prefix+"garbage")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("invalid key must get 401: got %d, body=%s", w.Code, w.Body.String())
	}
}
