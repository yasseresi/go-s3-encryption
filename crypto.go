package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var customerKey []byte // 32 bytes AES-256

// loadRuntimeConfig reads CUSTOMER_KEY_BASE64 and prepares the customerKey
func loadRuntimeConfig() error {
	b64 := os.Getenv("CUSTOMER_KEY_BASE64")
	if b64 == "" {
		return errors.New("CUSTOMER_KEY_BASE64 not set")
	}
	k, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}
	if len(k) != 32 {
		return errors.New("CUSTOMER_KEY_BASE64 must decode to 32 bytes (AES-256)")
	}
	customerKey = k
	return nil
}

// EncryptWithCustomerKey encrypts plaintext using AES-GCM and returns nonce||ciphertext bytes.
func EncryptWithCustomerKey(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(customerKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	// Prepend nonce so decrypt can extract it
	ct := gcm.Seal(nil, nonce, plaintext, nil)
	out := append(nonce, ct...)
	return out, nil
}

// DecryptWithCustomerKey expects nonce||ciphertext and returns plaintext
func DecryptWithCustomerKey(nonceAndCipher []byte) ([]byte, error) {
	block, err := aes.NewCipher(customerKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(nonceAndCipher) < ns {
		return nil, errors.New("ciphertext too short")
	}
	nonce := nonceAndCipher[:ns]
	ct := nonceAndCipher[ns:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, err
	}
	return pt, nil
}
