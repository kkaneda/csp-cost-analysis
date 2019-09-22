package main

import (
	"flag"
	costanalysis "github.com/kkaneda/csp-cost-analysis/pkg/costanalysis/aws"
	"log"
)

var (
	inputOfferFile string
	productsCSVFile string
	pricesCSVFile string
)

func init() {
	flag.StringVar(&inputOfferFile, "input-offer-file", "", "Input offer current index file")
	flag.StringVar(&productsCSVFile, "products-csv-file", "", "Output products CSV file")
	flag.StringVar(&pricesCSVFile, "prices-csv-file", "", "Output prices CSV file")

}

func main() {
	flag.Parse()
	if inputOfferFile == "" {
		log.Fatalf("--input-offer-file must be specified.")
	}
	if productsCSVFile == "" {
		log.Fatalf("--products-csv-file must be specified.")
	}
	if pricesCSVFile == "" {
		log.Fatalf("--prices-csv-file must be specified.")
	}

	o, err := costanalysis.ProcessOfferFile(inputOfferFile)
	if err != nil {
		log.Fatalf("failed to process offer file: %v", err)
	}
	if err := costanalysis.GenerateProductCSV(o, productsCSVFile); err != nil {
		log.Fatalf("failed to generate a products CSV file: %v", err)
	}
	if err := costanalysis.GeneratePricesCSV(o, pricesCSVFile); err != nil {
		log.Fatalf("failed to generate a prices CSV file: %v", err)
	}
}
