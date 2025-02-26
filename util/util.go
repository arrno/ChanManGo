package util

import (
	"math/rand/v2"
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
