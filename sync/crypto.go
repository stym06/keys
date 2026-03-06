package sync

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/scrypt"
)

var wordlist = []string{
	"alpha", "brave", "coral", "delta", "eagle", "flame", "ghost", "honey",
	"ivory", "jazz", "kite", "lemon", "maple", "noble", "olive", "pearl",
	"quilt", "river", "storm", "tiger", "urban", "vivid", "waltz", "xenon",
	"yacht", "zebra", "amber", "blade", "cedar", "drift", "ember", "frost",
	"grape", "haven", "inlet", "jewel", "knack", "lunar", "mango", "nexus",
	"ocean", "prism", "quest", "radar", "solar", "torch", "unity", "vault",
}

func GeneratePassphrase() (string, error) {
	words := make([]string, 3)
	for i := range words {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(wordlist))))
		if err != nil {
			return "", err
		}
		words[i] = wordlist[n.Int64()]
	}
	return fmt.Sprintf("%s-%s-%s", words[0], words[1], words[2]), nil
}

func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, 32)
}

func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
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

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	// prepend salt to ciphertext
	return append(salt, ciphertext...), nil
}

func Decrypt(data []byte, passphrase string) ([]byte, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("data too short")
	}

	salt := data[:16]
	ciphertext := data[16:]

	key, err := deriveKey(passphrase, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("data too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
