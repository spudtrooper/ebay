package ebay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/spudtrooper/goutil/html"
)

func HTML(query string) error {
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

	sort.Slice(items, func(a, b int) bool {
		ia, ib := items[a], items[b]
		return ia.Price > ib.Price
	})
	head := html.TableRowData{
		"PRICE",
		"TITLE",
		"BIDS",
	}
	var rows []html.TableRowData
	for _, it := range items {
		row := html.TableRowData{
			formatCurrency(it.Price),
			fmt.Sprintf(`<a href="%s">%s</a>`, it.URL, it.Title),
			fmt.Sprintf("%d", it.Bids),
		}
		rows = append(rows, row)
	}
	entities := []html.DataEntity{
		html.MakeDataEntityFromTable(html.TableData{
			Head: head,
			Rows: rows,
		}),
	}
	htmlData := html.Data{entities}

	html, err := html.Render(htmlData)
	if err != nil {
		return err
	}
	outFile := path.Join(dir, "index.html")
	if err := ioutil.WriteFile(outFile, []byte(html), 0755); err != nil {
		return err
	}

	log.Printf("wrote to %s", outFile)

	return nil
}
