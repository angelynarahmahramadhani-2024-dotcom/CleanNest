// Package security berisi utilitas keamanan yang dipakai lintas lapisan
// (usecase dan seeder repository), sehingga tidak bergantung pada framework apa pun.
package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

const (
	algorithm  = "pbkdf2_sha256"
	iterations = 100000
	saltLength = 16
)

// HashPassword menghasilkan hash PBKDF2-HMAC-SHA256 dengan salt acak per pengguna.
// Format: pbkdf2_sha256$<iterasi>$<salt base64>$<hash base64>
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return encode(password, salt, iterations), nil
}

// CheckPassword membandingkan password dengan hash tersimpan secara constant-time.
func CheckPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != algorithm {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(encode(password, salt, iter)), []byte(encoded)) == 1
}

func encode(password string, salt []byte, iter int) string {
	key := pbkdf2SHA256([]byte(password), salt, iter)
	return fmt.Sprintf("%s$%d$%s$%s", algorithm, iter,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

// pbkdf2SHA256 adalah implementasi PBKDF2 (RFC 8018) dengan panjang kunci 32 byte (1 blok SHA-256).
func pbkdf2SHA256(password, salt []byte, iter int) []byte {
	prf := hmac.New(sha256.New, password)
	blockIndex := make([]byte, 4)
	binary.BigEndian.PutUint32(blockIndex, 1)

	prf.Write(salt)
	prf.Write(blockIndex)
	u := prf.Sum(nil)

	result := make([]byte, len(u))
	copy(result, u)
	for i := 1; i < iter; i++ {
		prf.Reset()
		prf.Write(u)
		u = prf.Sum(u[:0])
		for j := range result {
			result[j] ^= u[j]
		}
	}
	return result
}
