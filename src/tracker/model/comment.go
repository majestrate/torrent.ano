package model

import (
	"time"
)

type Comment struct {
	ID     uint64
	Text   string
	Posted time.Time
}
