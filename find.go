package main

import (
	"flag"

	"github.com/spudtrooper/ebay/lib/ebay"
	"github.com/spudtrooper/goutil/check"
)

var (
	query           = flag.String("query", "", "item query, e.g. '1933 goudey'")
	seleniumVerbose = flag.Bool("selenium_verbose", false, "verbose selenium logging")
	seleniumHead    = flag.Bool("selenium_head", false, "Take screenshots withOUT headless chrome")
	page            = flag.Int("page", 0, "the only page to search")
	force           = flag.Bool("force", false, "Force writing, skipping the cache")
)

func main() {
	flag.Parse()
	check.CheckNonEmptyString(*query, "query")
	if err := ebay.Find(*query,
		ebay.FindPage(*page),
		ebay.FindSeleniumVerbose(*seleniumVerbose),
		ebay.FindSeleniumHead(*seleniumHead),
		ebay.FindForce(*force)); err != nil {
		panic(err.Error())
	}
}
