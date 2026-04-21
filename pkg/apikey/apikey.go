package apikey

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var base62Chars = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

func base62Encode(data []byte) string {
	num := new(big.Int).SetBytes(data)
	zero := big.NewInt(0)
	base := big.NewInt(62)
	mod := new(big.Int)
	var result []byte
	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append([]byte{base62Chars[mod.Int64()]}, result...)
	}
	if len(result) == 0 {
		return "0"
	}
	return string(result)
}

func base62Decode(s string) ([]byte, error) {
	num := new(big.Int)
	base := big.NewInt(62)
	for _, c := range []byte(s) {
		idx := -1
		for i, ch := range base62Chars {
			if ch == c {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, fmt.Errorf("invalid base62 character: %c", c)
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(idx)))
	}
	return num.Bytes(), nil
}

// Generate creates a new API key for the given user ID.
// Returns (plaintext_key, ciphertext_for_storage, error).
func Generate(secret []byte, userID uint, prefix string) (string, string, error) {
	block, err := aes.NewCipher(padKey(secret))
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	// payload: user_id (4 bytes) + timestamp (8 bytes) + random salt (16 bytes)
	payload := make([]byte, 28)
	binary.BigEndian.PutUint32(payload[0:4], uint32(userID))
	binary.BigEndian.PutUint64(payload[4:12], uint64(time.Now().Unix()))
	if _, err := rand.Read(payload[12:28]); err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, payload, nil) // nonce + encrypted + tag
	encoded := base62Encode(ciphertext)
	plainKey := prefix + encoded

	return plainKey, encoded, nil
}

// Validate decrypts the API key and returns the user ID.
func Validate(secret []byte, key string, prefix string) (uint, error) {
	if !strings.HasPrefix(key, prefix) {
		return 0, errors.New("invalid api key prefix")
	}
	encoded := key[len(prefix):]

	block, err := aes.NewCipher(padKey(secret))
	if err != nil {
		return 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	ciphertext, err := base62Decode(encoded)
	if err != nil {
		return 0, fmt.Errorf("base62 decode failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return 0, errors.New("ciphertext too short")
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	payload, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return 0, errors.New("decryption failed")
	}

	if len(payload) < 4 {
		return 0, errors.New("invalid payload")
	}

	userID := binary.BigEndian.Uint32(payload[0:4])
	return uint(userID), nil
}

func padKey(secret []byte) []byte {
	key := make([]byte, 32)
	copy(key, secret)
	return key
}
