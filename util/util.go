package util

import (
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"io"
	"math/big"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

func RandIdChars() string {
	lower := "abcdefghijklmnopqrstuvwxyz"
	upper := strings.ToUpper(lower)
	numeric := "0123456789"
	// special := "-_"
	return lower + upper + numeric
}

func Nanoid() string {
	str, err := gonanoid.Generate(RandIdChars(), 20)
	if err != nil {
		r := rand.Float64() * 100000000
		return strconv.FormatFloat(r, 'f', -1, 64)[:8]
	}
	return str
}

func InParallel[T any](funcs ...func() (T, error)) ([]T, error) {
	type orderedResult struct {
		result T
		order  int
	}
	dataChan := make(chan orderedResult, len(funcs))
	errChan := make(chan error, len(funcs))
	for i, f := range funcs {
		go func(order int) {
			result, err := f()
			if err != nil {
				errChan <- err
			} else {
				dataChan <- orderedResult{result, order}
			}
		}(i)
	}
	results := make([]T, len(funcs))
	for _ = range len(funcs) {
		select {
		case err := <-errChan:
			return nil, err
		case res := <-dataChan:
			results[res.order] = res.result
		}
	}
	return results, nil
}

func CreateHashedID(identifier string) string {
	// return identifier
	// Create a SHA-256 hash of the identifier
	hash := sha256.Sum256([]byte(identifier))

	// Define the character set: uppercase, lowercase, and all numbers (0-9)
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	base := len(charSet)

	// Convert the hash to an integer
	hashInt := new(big.Int).SetBytes(hash[:])

	// Create a 20-character identifier using the custom character set
	hashedID := make([]byte, 20)
	for i := 19; i >= 0; i-- {
		remainder := new(big.Int)
		hashInt.DivMod(hashInt, big.NewInt(int64(base)), remainder)
		hashedID[i] = charSet[remainder.Int64()]
	}

	return string(hashedID)
}

func IsValidName(name string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	return re.MatchString(name)
}

func GenerateRandomSecretKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := io.ReadFull(cryptoRand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
