package ratelimiter_test

import (
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/ratelimiter"
)

func TestInMemory_AllowsUpToBurstThenRejects(t *testing.T) {
	t.Parallel()

	// 0 refills/sec, burst 3: exactly 3 allowed, the 4th rejected.
	l := ratelimiter.NewInMemory(0, 3)

	for i := range 3 {
		if !l.Allow() {
			t.Fatalf("call %d should be allowed within burst", i+1)
		}
	}

	if l.Allow() {
		t.Fatal("call beyond burst should be rejected")
	}
}
