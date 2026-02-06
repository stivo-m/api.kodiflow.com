package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/stivo-m/api.kodiflow.com/pkg/config"
)

func loadEncryptionKey() ([]byte, error) {
	keyB64 := config.MustGetEnv("APP_ENCRYPTION_KEY")

	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, err
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key length: %d", len(key))
	}

	return key, nil
}

// Handles the encryption of a string with the apps's encryptoion key
func Encrypt(plaintext string, businessId string) (string, error) {
	key, err := loadEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), []byte(businessId))

	out := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Handles the decryption of a given string with the app's encryption key
func Decrypt(encrypted string, businessId string) (string, error) {
	key, err := loadEncryptionKey()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(businessId))
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
