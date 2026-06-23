package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

var errInvalidQueryParam = errors.New("invalid or missing query parameter")

func parseGenerateRequest(r *http.Request) (fizzbuzz.GenerateRequest, error) {
	q := r.URL.Query()

	int1, err := parseIntParam(q, "int1")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	int2, err := parseIntParam(q, "int2")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	limit, err := parseIntParam(q, "limit")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	str1, err := parseStringParam(q, "str1")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	str2, err := parseStringParam(q, "str2")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	return fizzbuzz.GenerateRequest{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}, nil
}

func parseIntParam(q url.Values, name string) (int, error) {
	raw := q.Get(name)
	if raw == "" {
		return 0, fmt.Errorf("%s is required: %w", name, errInvalidQueryParam)
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, errInvalidQueryParam)
	}

	return v, nil
}

func parseStringParam(q url.Values, name string) (string, error) {
	raw := q.Get(name)
	if raw == "" {
		return "", fmt.Errorf("%s is required: %w", name, errInvalidQueryParam)
	}

	return raw, nil
}
