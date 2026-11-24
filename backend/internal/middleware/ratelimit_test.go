package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterRefills(t *testing.T) {
	now := time.Unix(0, 0)
	rl := NewRateLimiter(2)
	rl.now = func() time.Time { return now }
	if !rl.Allow("ip") || !rl.Allow("ip") || rl.Allow("ip") {
		t.Fatal("capacidade de 2 por minuto não respeitada")
	}
	if !rl.Allow("outro-ip") {
		t.Fatal("chaves devem ser independentes")
	}
	now = now.Add(30 * time.Second) // 1 ficha reposta
	if !rl.Allow("ip") || rl.Allow("ip") {
		t.Fatal("reposição incorreta")
	}
	now = now.Add(2 * time.Minute)
	rl.Cleanup()
	if len(rl.buckets) != 0 {
		t.Fatal("baldes inativos deveriam ser removidos")
	}
}
