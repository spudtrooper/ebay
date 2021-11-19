package ebay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/yosssi/gohtml"
)

func HTML(query string, htmlOpts ...HTMLOption) error {
	opts := makeHTMLOptionImpl(htmlOpts...)

	c := makeCache(query)
	dir, err := c.dir()
	if err != nil {
		return err
	}

	var files []string
	if err := filepath.Walk(dir, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(f) == ".json" {
			files = append(files, f)
		}
		return nil
	}); err != nil {
		return err
	}

	var items []Item
	for _, f := range files {
		var its []Item
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &its); err != nil {
			return err
		}
		items = append(items, its...)
	}

	var output bytes.Buffer
	if err := outputHTML(&output, items, opts.InlineAssets()); err != nil {
		return err
	}

	html := gohtml.Format(output.String())
	outFile := path.Join(dir, "index.html")
	if err := ioutil.WriteFile(outFile, []byte(html), 0755); err != nil {
		return err
	}

	log.Printf("wrote to %s", outFile)

	return nil
}

func outputHTML(buf *bytes.Buffer, items []Item, inlineAssets bool) error {
	type tag string
	type attr struct {
		key, val string
	}
	const (
		trTag    tag = "tr"
		tdTag    tag = "td"
		thTag    tag = "th"
		tableTag tag = "table"
		theadTag tag = "thead"
		tbodyTag tag = "tbody"
	)
	out := func(ss ...string) {
		for _, s := range ss {
			buf.WriteString(s)
		}
		buf.WriteString("\n")
	}
	tagStart := func(t tag, attrs ...attr) {
		s := "<" + string(t)
		for _, at := range attrs {
			s += " " + at.key + "='" + at.val + "'"
		}
		s += ">"
		out(s)
	}
	tagEnd := func(t tag) {
		out("</" + string(t) + ">")
	}
	outputTag := func(t tag, ss ...string) {
		tagStart(t)
		if len(ss) > 0 {
			for _, s := range ss {
				out(s)
			}
			tagEnd(t)
		}
	}
	tr := func(s ...string) { outputTag(trTag, s...) }
	td := func(s ...string) { outputTag(tdTag, s...) }
	th := func(s string, attrs ...attr) {
		tagStart(thTag, append(attrs, attr{key: "class", val: "th-sm"})...)
		out(s)
		tagEnd(thTag)
	}
	table := func() {
		out(`<table class="sortable-table table table-striped table-bordered table-sm" cellspacing="0" width="100%">`)
	}
	pageStart := func() error {
		var css bytes.Buffer
		stylesheets := []string{
			"lib/third_party/mdb/css/mdb.lite.min.css",
			"lib/third_party/mdb/css/bootstrap.min.css",
			"lib/third_party/mdb/css/addons/datatables.min.css",
		}
		for _, f := range stylesheets {
			if inlineAssets {
				b, err := ioutil.ReadFile(f)
				if err != nil {
					return err
				}
				css.WriteString("<style>")
				css.WriteString(string(b))
				css.WriteString("</style>")
			} else {
				css.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="../%s"></link>`, f))
			}
			css.WriteString("\n")
		}

		var js bytes.Buffer
		javascripts := []string{
			"lib/third_party/mdb/js/jquery.min.js",
			"lib/third_party/mdb/js/mdb.min.js",
			"lib/third_party/mdb/js/bootstrap.min.js",
			"lib/third_party/mdb/js/addons/datatables.min.js",
		}
		for _, f := range javascripts {
			if inlineAssets {
				b, err := ioutil.ReadFile(f)
				if err != nil {
					return err
				}
				js.WriteString("<script>")
				js.WriteString(string(b))
				js.WriteString("</script>")
			} else {
				js.WriteString(fmt.Sprintf(`<script src="../%s"></script>`, f))
			}
			js.WriteString("\n")
		}

		out(`
	<!doctype html>
	<html lang="en">`)
		head, err := renderTemplate(`
	<head>
	<link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.1/css/all.min.css" rel="stylesheet"/>
	<link href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,700&display=swap" rel="stylesheet"/>	
	{{.Css}}
	{{.Js}}
	<script>
		$(document).ready(function () {
			$('.sortable-table').DataTable();
			$('.dataTables_length').addClass('bs-select');
		});			  
	</script>
	</head>
	`, "head", struct {
			Css, Js string
		}{
			Css: css.String(),
			Js:  js.String(),
		})
		if err != nil {
			return err
		}
		out(head)
		out("<body>")
		out(`<div class="container-fluid">`)
		out(`<a name="top"></a>`)
		return nil
	}

	pageEnd := func() {
		out("</div>")
		out("</body>")
		out("</html>")
	}

	outputItems := func() {
		table()
		tagStart(theadTag)
		tr()
		th("PRICE")
		th("TITLE")
		th("BIDS")
		tagEnd(trTag)
		tagEnd(theadTag)
		tagStart(tbodyTag)
		sort.Slice(items, func(a, b int) bool {
			ia, ib := items[a], items[b]
			return ia.Price > ib.Price
		})
		for _, it := range items {
			tr()
			td(formatCurrency(it.Price))
			td(fmt.Sprintf(`<a href="%s">%s</a>`, it.URL, it.Title))
			td(fmt.Sprintf("%d", it.Bids))
			tagEnd(trTag)
		}
		tagEnd(tbodyTag)
		tagEnd(tableTag)
	}

	if err := pageStart(); err != nil {
		return err
	}
	outputItems()
	pageEnd()

	return nil
}

func renderTemplate(t string, name string, data interface{}) (string, error) {
	tmpl, err := template.New(name).Parse(strings.TrimSpace(t))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
