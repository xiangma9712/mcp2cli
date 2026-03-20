package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"runtime"
)

// deriveKey produces a deterministic 32-byte AES key from the current
// executable path and OS/arch. This is obfuscation rather than strong
// encryption — it prevents casual reading of token files but does not
// protect against a determined attacker with access to the binary.
func deriveKey() []byte {
	h := sha256.New()

	// Executable path — changes per install location
	if exe, err := os.Executable(); err == nil {
		h.Write([]byte(exe))
	}

	// Runtime constants — ties key to specific build
	h.Write([]byte(runtime.GOOS))
	h.Write([]byte(runtime.GOARCH))

	// Fixed salt to namespace the key
	h.Write([]byte("mcp2cli-token-encryption-v1"))

	return h.Sum(nil)
}

func encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(deriveKey())
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(deriveKey())
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
