package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pimousse1099/fizz-buzz-api/internal/domain"
)

const testMaxLimit = 1000

func validReq() domain.GenerateFizzBuzzRequest {
	return domain.GenerateFizzBuzzRequest{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*domain.GenerateFizzBuzzRequest)
		wantErr bool
		// msgContains is checked only when wantErr is true.
		msgContains string
	}{
		{name: "valid", mutate: func(*domain.GenerateFizzBuzzRequest) {}, wantErr: false},
		{name: "limit at max is allowed", mutate: func(r *domain.GenerateFizzBuzzRequest) { r.Limit = testMaxLimit }, wantErr: false},
		{name: "int1 zero", mutate: func(r *domain.GenerateFizzBuzzRequest) { r.Int1 = 0 }, wantErr: true, msgContains: "int1"},
		{name: "int2 zero", mutate: func(r *domain.GenerateFizzBuzzRequest) { r.Int2 = 0 }, wantErr: true, msgContains: "int2"},
		{name: "limit zero", mutate: func(r *domain.GenerateFizzBuzzRequest) { r.Limit = 0 }, wantErr: true, msgContains: "limit"},
		{name: "limit above max", mutate: func(r *domain.GenerateFizzBuzzRequest) { r.Limit = testMaxLimit + 1 }, wantErr: true, msgContains: "limit"},
		{name: "str1 empty", mutate: func(r *domain.GenerateFizzBuzzRequest) { r.Str1 = "" }, wantErr: true, msgContains: "str1"},
		{name: "str2 empty", mutate: func(r *domain.GenerateFizzBuzzRequest) { r.Str2 = "" }, wantErr: true, msgContains: "str2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := validReq()
			tc.mutate(&req)

			err := req.Validate(testMaxLimit)

			check := assert.New(t)
			if !tc.wantErr {
				check.NoError(err)

				return
			}

			check.Error(err)
			check.ErrorIs(err, domain.ErrInvalidRequest)
			check.Contains(err.Error(), tc.msgContains)
		})
	}
}

// TestValidateErrorMessageOrder pins the message shape: the sentinel context
// comes first, the specific detail after.
func TestValidateErrorMessageOrder(t *testing.T) {
	t.Parallel()

	req := validReq()
	req.Limit = testMaxLimit + 1

	err := req.Validate(testMaxLimit)

	check := assert.New(t)
	check.Error(err)
	check.True(strings.HasPrefix(err.Error(), "failed to validate fizz-buzz request: "), err.Error())
	check.Contains(err.Error(), "limit must be between 1 and 1000, got 1001")
	check.True(errors.Is(err, domain.ErrInvalidRequest))
}
