package costanalysis

import (
	//"fmt"
	"encoding/json"
	"io/ioutil"
	"os"
	//	"strings"
	"testing"
)

func TestProcess(t *testing.T) {
	body, err := ioutil.ReadFile("testdata/offer-index-file.json")
	if err != nil {
		t.Fatal(err)
	}
	oif := OfferIndexFile{}
	err = json.Unmarshal(body, &oif)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcess2(t *testing.T) {
	r, err := os.Open("testdata/offer-current-index-file.json")
	if err != nil {
		t.Fatal(err)
	}
	// Took 13.07s
	o, err := process(r)
	if err != nil {
		t.Fatal(err)
	}

	if err := validate(o); err != nil {
		t.Error(err)
	}
}

func TestCompare(t *testing.T) {
	r, err := os.Open("testdata/offer-current-index-file.json")
	if err != nil {
		t.Fatal(err)
	}
	// Took 13.07s
	o, err := process(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := compareOnDemandPrice(o); err != nil {
		t.Fatal(err)
	}
}
