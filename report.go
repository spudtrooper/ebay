package main

import (
	"flag"

	"github.com/spudtrooper/ebay/lib/ebay"
	"github.com/spudtrooper/goutil/check"
)

var (
	query = flag.String("query", "", "item query, e.g. '1933 goudey'")
)

func main() {
	flag.Parse()
	check.CheckNonEmptyString(*query, "query")
	if err := ebay.HTML(*query); err != nil {
		panic(err.Error())
	}
}
