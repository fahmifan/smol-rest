package model

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func NewID() ulid.ULID {
	seed := time.Now()
	src := rand.NewSource(seed.UnixNano())
	entropy := ulid.Monotonic(rand.New(src), 0)
	return ulid.MustNew(ulid.Timestamp(seed), entropy)
}
