package main

import (
	"flag"

	"github.com/spudtrooper/ebay/lib/ebay"
	"github.com/spudtrooper/goutil/check"
)

var (
	query            = flag.String("query", "", "item query, e.g. '1933 goudey'")
	inlineHtmlAssets = flag.Bool("inline_html_assets", false, "Whether to inline th contents of CSS and JS files into HTML files")
)

func main() {
	flag.Parse()
	check.CheckNonEmptyString(*query, "query")
	if err := ebay.HTML(*query, ebay.HTMLInlineAssets(*inlineHtmlAssets)); err != nil {
		panic(err.Error())
	}
}
