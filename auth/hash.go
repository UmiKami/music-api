package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)

	if err != nil {
		return nil, err
	}

	return salt, nil
}

func HashPassword(password string) (string, error) {
	saltLength, err := strconv.Atoi(os.Getenv("SALT_LENGTH")); if err != nil {
		log.Fatal("Error while converting salt length to int")
	}

	salt, err := generateSalt(saltLength)

    if err != nil {
        log.Fatal(err)
    }

	memory := uint32(64 * 1024) // 64 MB
	time := uint32(3)           // 3 iterations
	threads := uint8(4)         // 4 parallel threads
	keyLength := uint32(32)     // 32 bytes

	// Generate hash using Argon2id
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLength)

	// Encode salt and hash to base64 and concatenate them
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s:%s", saltBase64, hashBase64), nil
}

func VerifyPassword(password, hashedPassword string) (bool, error) {
    parts := strings.Split(hashedPassword, ":")
    if len(parts) != 2 {
        return false, errors.New("invalid hash format")
    }

    salt, err := base64.RawStdEncoding.DecodeString(parts[0])
    if err != nil {
        return false, err
    }

    hash, err := base64.RawStdEncoding.DecodeString(parts[1])
    if err != nil {
        return false, err
    }

    memory := uint32(64 * 1024) // 64 MB
    time := uint32(3)           // 3 iterations
    threads := uint8(4)         // 4 parallel threads
    keyLength := uint32(32)     // 32 bytes

    testHash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLength)

    return string(testHash) == string(hash), nil
}
