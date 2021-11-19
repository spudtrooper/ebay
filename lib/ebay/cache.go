package ebay

import (
	"strings"

	"github.com/spudtrooper/goutil/io"
)

const (
	cacheDir = "data"
)

type cache struct {
	query string
}

func makeCache(query string) *cache {
	return &cache{query}
}

func (c *cache) dir() (string, error) {
	return io.MkdirAll(cacheDir, strings.ReplaceAll(c.query, " ", "-"))
}
