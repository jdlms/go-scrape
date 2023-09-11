package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

// defining a data structure to store the scraped data
type car struct {
	title, price, monthly, km, apr string
}

// it verifies if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func main() {
	// initializing the slice of structs that will contain the scraped data
	var cars []car

	// initializing the list of pages to scrape with an empty slice
	var pagesToScrape []string

	// the first pagination URL to scrape
	pageToScrape := "https://www.autotrader.ca/cars/on/toronto/?rcp=15&rcs=0&srt=35&prx=250&prv=Ontario&loc=Toronto%2C%20ON&hprc=True&wcp=True&inMarket=advancedSearch"

	// initializing the list of pages discovered with a pageToScrape
	pagesDiscovered := []string{pageToScrape}

	// current iteration
	i := 1
	// max pages to scrape
	limit := 5

	// initializing a Colly instance
	c := colly.NewCollector()
	// setting a valid User-Agent header
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	// iterating over the list of pagination links to implement the crawling logic
	c.OnHTML("a.page-numbers", func(e *colly.HTMLElement) {
		// discovering a new page
		newPaginationLink := e.Attr("href")

		// if the page discovered is new
		if !contains(pagesToScrape, newPaginationLink) {
			// if the page discovered should be scraped
			if !contains(pagesDiscovered, newPaginationLink) {
				pagesToScrape = append(pagesToScrape, newPaginationLink)
			}
			pagesDiscovered = append(pagesDiscovered, newPaginationLink)
		}
	})

	// scraping the product data
	c.OnHTML("div.result-item", func(e *colly.HTMLElement) {
		// Initialize a new car struct
		car := car{}

		// Title
		car.title = e.ChildText("span.result-title.click > span.title-with-trim")

		car.price = e.ChildText("div.price-amount > span.price-amount-value")
		// Kilometers
		kmText := e.ChildText("span.odometer-proximity")
		car.km = strings.TrimSpace(kmText)


		// Monthly Payment and APR
		car.monthly = e.ChildText("div.price-delta > div.price-outer-div > div > span.price-amount")
		car.apr = e.ChildText("div.price-delta > div.flex-center > div.price-delta-text > p")

		// Append the car data to the slice
		cars = append(cars, car)
	})

	// c.OnHTML("a.dealer-split-wrapper", func(e *colly.HTMLElement) {

	// 	// Price and Save Amount
	// 	car.price = e.ChildText("span.price-amount-value")

	// 	// Monthly Payment and APR
	// 	car.monthly = e.ChildText("span.payment-tag-installment")
	// 	car.apr = e.ChildText("span.payment-tag-rate")
	// })

	c.OnScraped(func(response *colly.Response) {
		// until there is still a page to scrape
		if len(pagesToScrape) != 0 && i < limit {
			// getting the current page to scrape and removing it from the list
			pageToScrape = pagesToScrape[0]
			pagesToScrape = pagesToScrape[1:]

			// incrementing the iteration counter
			i++

			// visiting a new page
			c.Visit(pageToScrape)
		}
	})

	// visiting the first page
	c.Visit(pageToScrape)

	// opening the CSV file
	file, err := os.Create("products.csv")
	if err != nil {
		log.Fatalln("Failed to create output CSV file", err)
	}
	defer file.Close()

	// initializing a file writer
	writer := csv.NewWriter(file)

	// defining the CSV headers
	headers := []string{
		"title",
		"price",
		"km",
		"monthly",
		"apr",
	}
	// writing the column headers
	writer.Write(headers)

	// adding each Pokemon product to the CSV output file
	for _, car := range cars {
		// converting a PokemonProduct to an array of strings
		record := []string{
			car.title,
			car.price,
			car.km,
			car.monthly,
			car.apr,
		}

		// writing a new CSV record
		writer.Write(record)
	}
	defer writer.Flush()
}

// type car struct {
// 	title, price, km, monthly, apr string
// }
