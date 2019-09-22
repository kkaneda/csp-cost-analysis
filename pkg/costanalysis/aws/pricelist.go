package costanalysis

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"

	//"fmt"
	"io/ioutil"
	"net/http"
)

// AWS Price List API (https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/using-ppslong.html)
//
// https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/reading-an-offer.html

type Offer struct {
	OfferCode             string `json:"offerCode"`
	VersionIndexURL       string `json:"versionIndexUrl"`
	CurrentVersionURL     string `json:"currentVersionUrl"`
	CurrentRegionIndexURL string `json:"currentRegionIndexUrl"`
}

type OfferIndexFile struct {
	FormatVersion   string           `json:"formatVersion"`
	Disclaimer      string           `json:"disclaimer"`
	PublicationDate string           `json:"publicationDate"`
	Offers          map[string]Offer `json:"offers"`
}

type Attributes struct {
	ServiceCode       string `json:"servicecode"`
	Location          string `json:"location"`
	LocationType      string `json:"locationType"`
	InstanceType      string `json:"instanceType"`
	CurrentGeneration string `json:"currentGeneration"`
	InstanceFamily    string `json:"instanceFamily"`
	VCPU              string `json:"vcpu"`
	PhysicalProcessor string `json:"physicalProcessor"`
	ClockSpeed        string `json:"clockSpeed"`
	Memory            string `json:"memory"`

	Storage               string `json:"storage"`
	NetworkPerformance    string `json:"networkPerformance"`
	ProcessorArchitecture string `json:"processorArchitecture"`
	// "Shared", "Host" or "Dedicated"
	Tenancy string `json:"tenancy"`

	OperatingSystem string `json:"operatingSystem"`

	LicenseModel   string `json:"licenseModel"`
	CapacityStatus string `json:"capacitystatus"`

	PreInstalledSw string `json:"preInstalledSw"`
}

type Product struct {
	SKU        string     `json:"sku"`
	Attributes Attributes `json:"attributes"`
}

type PricePerUnit struct {
	RateCode     string            `json:"rateCode"`
	BeginRange   string            `json:"beginRange"`
	EndRange     string            `json:"endRange"`
	Unit         string            `json:"unit"`
	PricePerUnit map[string]string `json:"pricePerUnit"`
}

// For each SKU, following six
const (
	OnDemandOfferTermCode = "JRTCKXETXF"

	/*
	   Here are all possible offer terms. For some reason, there is a dupe.

	      Z2E3P23VKM. {LeaseContractLength:3yr OfferingClass:convertible PurchaseOption:No Upfront}
	      R5XV2EPZQZ. {LeaseContractLength:3yr OfferingClass:convertible PurchaseOption:Partial Upfront}
	      MZU6U2429S. {LeaseContractLength:3yr OfferingClass:convertible PurchaseOption:All Upfront}

	      UDM74VY9CQ. {LeaseContractLength:3 yr OfferingClass:standard PurchaseOption:NoUpfront}
	      3RJ4P9STGK. {LeaseContractLength:3 yr OfferingClass:standard PurchaseOption:PartialUpfront}
	      CPPW92X9U4. {LeaseContractLength:3 yr OfferingClass:standard PurchaseOption:AllUpfront}

	      // What's difference between this and UDM74VY9CQ?
	      BPH4J8HBKS. {LeaseContractLength:3yr OfferingClass:standard PurchaseOption:No Upfront}
	      NQ3QZPMQV9. {LeaseContractLength:3yr OfferingClass:standard PurchaseOption:All Upfront}
	      38NPMPTW36. {LeaseContractLength:3yr OfferingClass:standard PurchaseOption:Partial Upfront}

	      7NE97W5U4E. {LeaseContractLength:1yr OfferingClass:convertible PurchaseOption:No Upfront}
	      CUZHX8X6JH. {LeaseContractLength:1yr OfferingClass:convertible PurchaseOption:Partial Upfront}
	      VJWZNREJX2. {LeaseContractLength:1yr OfferingClass:convertible PurchaseOption:All Upfront}

	      4NA7Y494T4. {LeaseContractLength:1yr OfferingClass:standard PurchaseOption:No Upfront}
	      HU7G6KETJZ. {LeaseContractLength:1yr OfferingClass:standard PurchaseOption:Partial Upfront}
	      6QCMYABX3D. {LeaseContractLength:1yr OfferingClass:standard PurchaseOption:All Upfront}

	      CRSZ9MHARF. {LeaseContractLength:1 yr OfferingClass:standard PurchaseOption:NoUpfront}
	      YKSHTEAGQM. {LeaseContractLength:1 yr OfferingClass:standard PurchaseOption:PartialUpfront}
	      DXEJ4EGHUJ. {LeaseContractLength:1 yr OfferingClass:standard PurchaseOption:AllUpfront}
	*/
)

const (
	PricePerHourPriceTermCode = "6YS6EN2CT7"
	UpfrontFeePriceTermCode   = "2TG2D8R56U"
)

type TermAttributes struct {
	// "1yr", "1 yr", "3yr", or "3 yr"
	LeaseContractLength string `json:"LeaseContractLength"`

	// "standard" or "convertible".
	OfferingClass string `json:"OfferingClass"`

	// "No Upfront", "Partial Upfront", "All Upfront",
	// "NoUpfront", "PartialUpfront", or "AllUpfront"
	PurchaseOption string `json:"PurchaseOption"`
}

type Term struct {
	// SKU is a unique code for a product.
	SKU string `json:"sku"`

	// OfferTermCode is a unique code for a specific type of term.
	// For example, KCAKZHGHG. Product and price combinations are referenced
	// by the SKU code followed by the term code, separated by a period.
	// For example, U7ADXS4BEK5XXHRU.KCAKZHGHG.
	OfferTermCode string `json:"offerTermCode"`

	// EffectiveDate is the date that an offer file goes into effect.
	// For example, if a term has an EffectiveDate of November 1, 2017,
	// the price is not valid before November 1, 2017.
	EffectiveDate string `json:"effectiveDate"`

	// PriceDimensions is the pricing details for the offer file,
	// such as how usage is measured, the currency that you can use to pay with,
	// and the pricing tier limitations.
	PriceDimensions map[string]PricePerUnit `json:"priceDimensions"`

	TermAttributes TermAttributes `json:"termAttributes"`
}

type OfferCurrentVersionIndexFile struct {
	FormatVersion   string             `json:"formatVersion"`
	Disclaimer      string             `json:"disclaimer"`
	OfferCode       string             `json:"offerCode"`
	Version         string             `json:"version"`
	PublicationDate string             `json:"publicationDate"`
	Products        map[string]Product `json:"products"`
	// The specific type of term that a term definition describes.
	// The valid term types are reserved and onDemand.
	Terms map[string]map[string]map[string]Term `json:"terms"`
}

// downloadFile downloads a file from a specified URL.
func downloadFile(URL string) ([]byte, error) {
	resp, err := http.Get(URL)
	if err != nil {
		// TODO: do chaining
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}

func processOfferIndexFile() (*OfferCurrentVersionIndexFile, error) {
	domainName := "https://pricing.us-east-1.amazonaws.com"
	// Download the offer index file
	URL := domainName + "/offers/v1.0/aws/index.json"
	body, err := downloadFile(URL)
	if err != nil {
		return nil, err
	}

	oif := OfferIndexFile{}
	err = json.Unmarshal(body, &oif)
	if err != nil {
		return nil, err
	}

	offer, ok := oif.Offers["AmazonEC2"]
	if !ok {
		return nil, fmt.Errorf("'AmazonEC2' not found in the offer index file")
	}

	URL = domainName + offer.CurrentRegionIndexURL
	resp, err := http.Get(URL)
	if err != nil {
		// TODO: wrap an error
		return nil, err
	}
	return process(resp.Body)
}

func process(r io.Reader) (*OfferCurrentVersionIndexFile, error) {
	ocvifs := []*OfferCurrentVersionIndexFile{}

	dec := json.NewDecoder(r)
	for dec.More() {
		ocvif := OfferCurrentVersionIndexFile{}
		err := dec.Decode(&ocvif)
		if err != nil {
			// TODO: wrap an error
			return nil, err
		}
		ocvifs = append(ocvifs, &ocvif)
	}

	if len(ocvifs) != 1 {
		return nil, fmt.Errorf("expected only one entry in the json file, but got %d entries", len(ocvifs))
	}

	return ocvifs[0], nil
}

func ProcessOfferFile(filename string) (*OfferCurrentVersionIndexFile, error) {
	r, err := os.Open(filename)
	defer r.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to open an offer json file: %s", filename)
	}
	return process(r)
}

func validate(o *OfferCurrentVersionIndexFile) error {
	offerTermCodes := map[string]TermAttributes{}

	for _, v := range o.Terms {
		for sku, vv := range v {
			_, ok := o.Products[sku]
			if !ok {
				return fmt.Errorf("not found")
			}

			for _, term := range vv {
				tas, ok := offerTermCodes[term.OfferTermCode]
				if ok {
					if !reflect.DeepEqual(tas, term.TermAttributes) {
						return fmt.Errorf("term attributes mismatch: %+v v.s %+v",
							tas, term.TermAttributes)
					}
				} else {
					offerTermCodes[term.OfferTermCode] = term.TermAttributes
				}
			}
		}
	}

	for k, v := range offerTermCodes {
		fmt.Printf("%v. %+v\n", k, v)
	}

	return nil
}

func getOnDemandPrices(o *OfferCurrentVersionIndexFile) (map[string]*Term, error) {
	result := map[string]*Term{}

	for _, v := range o.Terms {
		for sku, vv := range v {
			for _, term := range vv {
				if term.OfferTermCode != OnDemandOfferTermCode {
					continue
				}
				if _, ok := result[sku]; ok {
					return nil, fmt.Errorf("dup")
				}

				result[sku] = &term
			}
		}
	}
	return result, nil
}

func GenerateProductCSV(o *OfferCurrentVersionIndexFile, outputFilename string) error {
	f, err  := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("cannot create an ouput CSV file: %s", outputFilename);
	}
	defer f.Close()

	w := csv.NewWriter(f)

	for sku, product := range o.Products {
		attrs := product.Attributes
		r := []string{
			sku,
			attrs.InstanceType,
			attrs.InstanceFamily,
			attrs.Storage,
			attrs.Tenancy,
			attrs.OperatingSystem,
			attrs.LicenseModel,
			attrs.CapacityStatus,
			attrs.PreInstalledSw,
			attrs.Location,
		}
		if err := w.Write(r); err != nil {
			return err
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("failed to write a CSV file: %v", err);
	}
	return err;
}


func GeneratePricesCSV(o *OfferCurrentVersionIndexFile, outputFilename string) error {
	f, err  := os.Create(outputFilename)
	if err != nil {
		return fmt.Errorf("cannot create an ouput CSV file: %s", outputFilename);
	}
	defer f.Close()

	w := csv.NewWriter(f)

	i := 0
	for _, v := range o.Terms {
		for sku, vv := range v {
			for _, term := range vv {
				pds := term.PriceDimensions
				a := term.TermAttributes
				for _, pd := range pds {
					usd, ok := pd.PricePerUnit["USD"]
					if !ok {
						fmt.Printf("no USD price tag for sku %s\n", sku)
						continue
					}

					r := []string{
						fmt.Sprintf("%d", i),
						sku,
						term.OfferTermCode,
						term.EffectiveDate,
						pd.RateCode,
						pd.BeginRange,
						pd.EndRange,
						pd.Unit,
						usd,
						a.LeaseContractLength,
						a.OfferingClass,
						a.PurchaseOption,
					}
					if err := w.Write(r); err != nil {
						return err
					}
					i += 1
				}
			}
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("failed to write a CSV file: %v", err);
	}
	return err;
}

func compareOnDemandPrice(o *OfferCurrentVersionIndexFile) error {
	// Create a map: Instance type -> (SKU, location, price).

	type InstanceRecord struct {
		sku      string
		location string
	}

	m := map[string][]*InstanceRecord{}

	for sku, product := range o.Products {
		instType := product.Attributes.InstanceType
		loc := product.Attributes.Location

		recs, ok := m[instType]
		if !ok {
			recs = []*InstanceRecord{}
		}

		i := InstanceRecord{
			sku:      sku,
			location: loc,
		}
		m[instType] = append(recs, &i)
	}

	prices, err := getOnDemandPrices(o)
	if err != nil {
		// TODO: chaining
		return err
	}

	// Two issues:
	// - We see SKU with no associated on demand price info
	// - We see SKU whose price is 0 USD?

	for instType, recs := range m {
		lowest, lowestLoc, lowestSku := -1.0, "", ""
		highest, highestLoc, highestSku := -1.0, "", ""

		for _, rec := range recs {

			price, ok := prices[rec.sku]
			if !ok {
				fmt.Printf("no price found for sku %s\n", rec.sku)
				continue
			}

			if instType == "m5.2xlarge" && rec.location == "Asia Pacific (Seoul)" {
				fmt.Printf("sku = %s, price=%v\n", rec.sku, price)
			}

			for _, pd := range price.PriceDimensions {
				usd, ok := pd.PricePerUnit["USD"]
				if !ok {
					fmt.Printf("no USD price tag for sku %s\n", rec.sku)
					continue
				}
				p, err := strconv.ParseFloat(usd, 64)
				if err != nil {
					return err
				}
				if lowest < 0.0 || p < lowest {
					lowest = p
					lowestLoc = rec.location
					lowestSku = rec.sku
				}
				if highest < 0.0 || p > highest {
					highest = p
					highestLoc = rec.location
					highestSku = rec.sku
				}
			}
		}
		fmt.Printf("instType: %s, len(recs) = %d, highest=%v (%v, %v), lowest=%v (%v, %v), diff=%v\n",
			instType, len(recs), highest, highestLoc, highestSku, lowest, lowestLoc, lowestSku, highest-lowest)
	}
	return nil
}
