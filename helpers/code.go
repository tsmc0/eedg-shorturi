package helpers

import (
	"crypto/rand"
	"encoding/binary"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// EncodeBase62 кодирует uint64 в base62-строку минимальной длины.
func EncodeBase62(n uint64) string {
	if n == 0 {
		return string(base62Alphabet[0])
	}
	var buf [11]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = base62Alphabet[n%62]
		n /= 62
	}
	return string(buf[i:])
}

// RandomUint64 возвращает криптослучайное число — используется как seed для ID.
func RandomUint64() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b[:]), nil
}

// NewCode генерирует base62-код длиной не меньше minLen.
// Т.к. EncodeBase62 не сохраняет лидирующие нули, короткие коды добиваются
// перегенерацией через случайный seed.
func NewCode(minLen int) (string, error) {
	for {
		n, err := RandomUint64()
		if err != nil {
			return "", err
		}
		code := EncodeBase62(n)
		if len(code) >= minLen {
			return code, nil
		}
	}
}
