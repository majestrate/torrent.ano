package index

import (
	"errors"
	"html"
	"regexp"
	"strings"
)

func CheckTags(tags string) error {
	const maxTagNameSize int = 60
	tags = html.EscapeString(tags) //paranoic mode
	tags_ := strings.Split(tags, ",")
	tagsRe := regexp.MustCompile("^[a-z0-9]*$")

	for i := len(tags_) - 1; i >= 0; i-- {
		for i1 := len(tags_) - 1; i1 > 0; i1-- {
			if i == i1 {
				continue
			}
			if tags_[i] == tags_[i1] {
				return errors.New("Tags error - " + tags_[i] + " exists already")
			}
		}
		if len(tags_[i]) > maxTagNameSize {
			return errors.New("Tags error - " + tags_[i] + " is big size name")
		}
		if !tagsRe.MatchString(tags_[i]) {
			return errors.New("Tags error - " + tags_[i] + " unallowed chars")
		}
	}
	return nil

}
