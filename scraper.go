package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Car struct {
	name, price, dealer, description, status string
}

func scrapePages(startPage, endPage int, carsCh chan<- []Car, wg *sync.WaitGroup) {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[3:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	defer wg.Done()

	parentCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	var cars []Car
	var nodes []*cdp.Node

	// Navigate and collect car nodes
	if err := chromedp.Run(ctx,
		chromedp.Navigate(`https://www.autotrader.ca/cars/on/toronto/?rcp=15&rcs=0&srt=35&prx=250&prv=Ontario&loc=Toronto%2C%20ON&hprc=True&wcp=True&inMarket=advancedSearch`),
		chromedp.Sleep(2*time.Second),
		chromedp.Nodes(`a.inner-link`, &nodes, chromedp.ByQueryAll),
	); err != nil {
		log.Printf("Error running chromedp tasks: %v", err)
		return
	}

	for _, node := range nodes {
		var name, price, dealer, description, status string

		if err := chromedp.Run(ctx,
			chromedp.WaitVisible(`a.inner-link`, chromedp.FromNode(node)),
			chromedp.Click(`a.inner-link`, chromedp.FromNode(node)),
			chromedp.Text(`h1.hero-title`, &name),
			chromedp.Text(`p.hero-price`, &price),
			chromedp.Text(`p.col-xs-12.no-padding.di-name`, &dealer),
			chromedp.Text(`div#vdp-collapsible-short-text`, &description),
			chromedp.Text(`span#spec-value-1`, &status),
			chromedp.Navigate("javascript:window.history.back();"),
		); err != nil {
			log.Printf("Error clicking and extracting details: %v", err)
			continue
		}

		cars = append(cars, Car{name, price, dealer, description, status})
	}

	carsCh <- cars

	if err := chromedp.Run(ctx, chromedp.Click(`a.page-link-2`, chromedp.NodeVisible)); err != nil {
		log.Printf("Error clicking next page: %v", err)
	}
}

func main() {
	carsCh := make(chan []Car)
	var wg sync.WaitGroup

	wg.Add(1)
	go scrapePages(1, 2, carsCh, &wg)

	go func() {
		wg.Wait()
		close(carsCh)
	}()

	var allCars []Car
	for cars := range carsCh {
		allCars = append(allCars, cars...)
	}

	file, err := os.Create("cars.csv")
	if err != nil {
		log.Fatalf("Could not create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Name", "Price", "Dealer", "Description", "Status"}
	if err := writer.Write(headers); err != nil {
		log.Printf("Error writing CSV header: %v", err)
		return
	}

	for _, car := range allCars {
		record := []string{car.name, car.price, car.dealer, car.description, car.status}
		if err := writer.Write(record); err != nil {
			log.Printf("Error writing CSV record: %v", err)
		}
	}
}
