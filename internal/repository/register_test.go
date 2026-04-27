package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/moontv/server/internal/database"
	"github.com/moontv/server/internal/model"
)

func setupRepoDB(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	database.Init(tmp + "/test.db")
}

func seedInvite(t *testing.T, expiresDays int) string {
	t.Helper()
	code := "ABCDE123"
	invite := model.InviteCode{
		Code:      code,
		CreatedBy: 1,
	}
	if expiresDays != 0 {
		exp := time.Now().AddDate(0, 0, expiresDays)
		invite.ExpiresAt = &exp
	}
	if err := database.DB.Create(&invite).Error; err != nil {
		t.Fatalf("seed invite: %v", err)
	}
	return code
}

func seedGlobalSource(t *testing.T, key string) {
	t.Helper()
	src := model.Source{
		Key:    key,
		Name:   "Test Source",
		APIUrl: "https://example.com/api",
	}
	if err := database.DB.Create(&src).Error; err != nil {
		t.Fatalf("seed source: %v", err)
	}
}

func TestRegisterUserSuccess(t *testing.T) {
	setupRepoDB(t)
	code := seedInvite(t, 0)
	seedGlobalSource(t, "testsrc")

	user, err := RegisterUser("alice", "hashedpassword", code)
	if err != nil {
		t.Fatalf("RegisterUser: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected user ID > 0")
	}
	if user.Username != "alice" {
		t.Fatalf("expected username alice, got %s", user.Username)
	}
	if user.Role != "user" {
		t.Fatalf("expected role user, got %s", user.Role)
	}

	// Verify invite marked used
	var invite model.InviteCode
	database.DB.Where("code = ?", code).First(&invite)
	if invite.UsedBy == nil || *invite.UsedBy != user.ID {
		t.Fatal("invite not marked as used")
	}

	// Verify global sources copied
	var sources []model.Source
	database.DB.Where("user_id = ?", user.ID).Find(&sources)
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
}

func TestRegisterUserRejectsUsedInvite(t *testing.T) {
	setupRepoDB(t)
	code := seedInvite(t, 0)

	userID := uint(999)
	usedAt := time.Now()
	database.DB.Model(&model.InviteCode{}).Where("code = ?", code).Updates(map[string]any{
		"used_by": userID,
		"used_at": usedAt,
	})

	_, err := RegisterUser("bob", "hash", code)
	if !errors.Is(err, ErrInviteUsed) {
		t.Fatalf("expected ErrInviteUsed, got: %v", err)
	}
}

func TestRegisterUserRejectsExpiredInvite(t *testing.T) {
	setupRepoDB(t)
	_ = seedInvite(t, -1)

	_, err := RegisterUser("charlie", "hash", "ABCDE123")
	if !errors.Is(err, ErrInviteExpired) {
		t.Fatalf("expected ErrInviteExpired, got: %v", err)
	}
}

func TestRegisterUserRejectsInvalidInvite(t *testing.T) {
	setupRepoDB(t)

	_, err := RegisterUser("dave", "hash", "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for nonexistent invite")
	}
}

func TestRegisterUserRejectsDuplicateUsername(t *testing.T) {
	setupRepoDB(t)
	code1 := seedInvite(t, 0)

	_, err := RegisterUser("eve", "hash", code1)
	if err != nil {
		t.Fatalf("first register: %v", err)
	}

	// Generate a second invite
	code2 := model.InviteCode{Code: "ZZZZZ999", CreatedBy: 1}
	database.DB.Create(&code2)

	_, err = RegisterUser("eve", "hash2", "ZZZZZ999")
	if !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("expected ErrUsernameExists, got: %v", err)
	}
}
