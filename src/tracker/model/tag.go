package model

import (
	"fmt"
)

type Tag struct {
	Name string
	ID   uint64
	Rank uint64
}

func (t *Tag) SearchLink() string {
	return fmt.Sprintf("/s/?id=%d", t.ID)
}
