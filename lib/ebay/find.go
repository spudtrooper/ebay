package ebay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/spudtrooper/goutil/io"
	goutilselenium "github.com/spudtrooper/goutil/selenium"
	"github.com/tebeka/selenium"
)

type Item struct {
	Title string ""
	URL   string
	Price float64
	Bids  int
}

func Find(query string, findOpts ...FindOption) error {
	opts := makeFindOptionImpl(findOpts...)

	wd, cancel, err := goutilselenium.MakeWebDriver(goutilselenium.MakeWebDriverOptions{
		Verbose:  opts.seleniumVerbose,
		Headless: !opts.seleniumHead,
	})
	if err != nil {
		return err
	}
	defer cancel()

	c := makeCache(query)
	dir, err := c.dir()
	if err != nil {
		return err
	}

	lookupPage := func(page int, useCache bool) (bool, int, error) {
		cacheFile := path.Join(dir, fmt.Sprintf("%d.json", page))

		if useCache && io.FileExists(cacheFile) {
			var items []Item
			b, err := ioutil.ReadFile(cacheFile)
			if err != nil {
				return false, 0, err
			}
			if err := json.Unmarshal(b, &items); err != nil {
				return false, 0, err
			}
			log.Printf("read %d items from from cache file %q", len(items), cacheFile)
			return false, len(items), nil
		}

		items, err := findHelper(wd, query, page)
		if err != nil {
			return true, 0, nil
		}
		log.Printf("caching %d items to %q", len(items), cacheFile)
		b, err := json.Marshal(&items)
		if err != nil {
			return false, 0, err
		}
		if err := ioutil.WriteFile(cacheFile, b, 0755); err != nil {
			return false, 0, err
		}

		sort.Slice(items, func(a, b int) bool {
			ia, ib := items[a], items[b]
			return ia.Price > ib.Price
		})

		for i, it := range items {
			log.Printf("[%d] %s: %s (%d bids)", i, formatCurrency(it.Price), it.Title, it.Bids)
		}

		return false, len(items), nil
	}

	if opts.page != 0 {
		_, _, err := lookupPage(opts.page, false)
		if err != nil {
			return err
		}
	} else {
		var lastLen int
		for page := 1; ; page++ {
			done, l, err := lookupPage(page, !opts.force)
			if err != nil {
				return err
			}
			if done || lastLen == l {
				break
			}
			lastLen = l
		}
	}

	return nil
}

func findHelper(wd selenium.WebDriver, query string, page int) ([]Item, error) {
	url := fmt.Sprintf("https://www.ebay.com/sch/i.html?_from=R40&_nkw=%s&_sacat=0&_sop=16&_ipg=200&_pgn=%d", strings.ReplaceAll(query, " ", "+"), page)
	log.Printf("looking up %s", url)

	if err := wd.Get(url); err != nil {
		return nil, err
	}

	var items []Item

	searchItemEl := func(it selenium.WebElement) error {
		titleEl, err := it.FindElement(selenium.ByClassName, "s-item__title")
		if err != nil {
			return err
		}
		title, err := titleEl.Text()
		if err != nil {
			return err
		}
		if title == "" {
			return nil
		}
		if strings.HasPrefix(title, "NEW LISTING") {
			title = strings.Replace(title, "NEW LISTING", "", 1)
		}

		itemLinkEl, err := it.FindElement(selenium.ByClassName, "s-item__link")
		if err != nil {
			return err
		}
		url, err := itemLinkEl.GetAttribute("href")
		if err != nil {
			return err
		}

		priceEl, err := it.FindElement(selenium.ByClassName, "s-item__price")
		if err != nil {
			return err
		}
		priceStr, err := priceEl.Text()
		if err != nil {
			return err
		}
		price, err := convertToPrice(priceStr)
		if err != nil {
			return err
		}

		var bids int
		bidEl, err := it.FindElement(selenium.ByClassName, "s-item__bids")
		if err == nil && bidEl != nil {
			bidsStr, err := bidEl.Text()
			if err != nil {
				return err
			}
			bidsStr = strings.Replace(bidsStr, " bids", "", 1)
			b, err := strconv.Atoi(bidsStr)
			if err != nil {
				return err
			}
			bids = b
		}

		item := Item{
			Title: title,
			URL:   url,
			Price: price,
			Bids:  bids,
		}
		items = append(items, item)

		return nil
	}

	findItems := func() ([]Item, error) {
		itemEls, err := wd.FindElements(selenium.ByClassName, "s-item")
		if err != nil {
			return nil, err
		}
		log.Printf("found %d items", len(itemEls))
		for _, itemEl := range itemEls {
			if err := searchItemEl(itemEl); err != nil {
				return nil, err
			}
		}
		return items, nil
	}
	wd.Wait(func(wd selenium.WebDriver) (bool, error) {
		its, err := findItems()
		if err != nil {
			return false, err
		}
		if len(its) == 0 {
			return false, nil
		}
		items = its
		return true, nil
	})

	return items, nil
}

func formatCurrency(f float64) string {
	return fmt.Sprintf("$%0.2f", f)
}

func convertToPrice(s string) (float64, error) {
	// $80.70
	s = strings.ReplaceAll(s, "$", "")
	// $123,345.89
	s = strings.ReplaceAll(s, ",", "")
	// $80.70 to $142.00
	s = strings.Split(s, " ")[0]
	return strconv.ParseFloat(s, 64)
}
