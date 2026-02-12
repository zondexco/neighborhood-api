package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	argon2MemoryKB uint32 = 64 * 1024
	argon2Time     uint32 = 3
	argon2Threads  uint8  = 2
	argon2KeyLen   uint32 = 32
	argon2SaltLen  uint32 = 16
)

func HashPIN(pin string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(pin), salt, argon2Time, argon2MemoryKB, argon2Threads, argon2KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argon2MemoryKB, argon2Time, argon2Threads, b64Salt, b64Hash), nil
}

// VerifyPIN valida el PIN contra hash almacenado.
// Retorna: valid, needsRehash, error
func VerifyPIN(pin, encodedHash string) (bool, bool, error) {
	if strings.HasPrefix(encodedHash, "$argon2id$") {
		valid, err := verifyArgon2id(pin, encodedHash)
		return valid, false, err
	}

	if strings.HasPrefix(encodedHash, "$2a$") || strings.HasPrefix(encodedHash, "$2b$") || strings.HasPrefix(encodedHash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(pin))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return false, false, nil
			}
			return false, false, fmt.Errorf("failed to verify bcrypt pin: %w", err)
		}
		return true, true, nil
	}

	return false, false, fmt.Errorf("unsupported password hash format")
}

func verifyArgon2id(pin, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid argon2 hash format")
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid argon2 algorithm")
	}

	if parts[2] != "v=19" {
		return false, fmt.Errorf("unsupported argon2 version")
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false, fmt.Errorf("invalid argon2 params")
	}

	memory, err := parseArgonParam(params[0], "m")
	if err != nil {
		return false, err
	}
	timeCost, err := parseArgonParam(params[1], "t")
	if err != nil {
		return false, err
	}
	threads, err := parseArgonParam(params[2], "p")
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid argon2 salt encoding: %w", err)
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid argon2 hash encoding: %w", err)
	}

	calculatedHash := argon2.IDKey([]byte(pin), salt, uint32(timeCost), uint32(memory), uint8(threads), uint32(len(storedHash)))
	if subtle.ConstantTimeCompare(storedHash, calculatedHash) == 1 {
		return true, nil
	}

	return false, nil
}

func parseArgonParam(input, key string) (int, error) {
	prefix := key + "="
	if !strings.HasPrefix(input, prefix) {
		return 0, fmt.Errorf("invalid argon2 %s param", key)
	}

	value, err := strconv.Atoi(strings.TrimPrefix(input, prefix))
	if err != nil {
		return 0, fmt.Errorf("invalid argon2 %s value: %w", key, err)
	}

	return value, nil
}
