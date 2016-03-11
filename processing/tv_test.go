package processing

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestParseInfo(t *testing.T) {
	expectations := map[string]*TVNameParseResult{
		"Sleepy.Hollow.S03E01.720p.HDTV.x264-AVS": {
			Name:        "Sleepy Hollow",
			CleanedName: "Sleepy Hollow S03E01 720p HDTV x264-AVS",
			Country:     "",
			Episode:     "1",
			Season:      "3",
			Airdate:     "",
		},
		"NBC.Nightly.News.2016.02.17.WEB-DL.x264-2Maverick": {
			Name:        "NBC Nightly News",
			CleanedName: "NBC Nightly News 2016 02 17 WEB-DL x264-2Maverick",
			Country:     "",
			Episode:     "02/17",
			Season:      "2016",
			Airdate:     "2016.02.17",
		},
	}
	for name, v := range expectations {
		res, err := ParseInfo(name)
		if err != nil {
			t.Fatalf("Error %v", err)
		}
		if !reflect.DeepEqual(v, res) {
			t.Errorf("Diff in Parse output for %s", name)
			fmt.Println("Expected:")
			spew.Dump(v)
			fmt.Println("Got:")
			spew.Dump(res)
		}
	}
}

func TestParseNameToTVName(t *testing.T) {
	expectations := [][]string{
		{"Sleepy.Hollow.S03E01.720p.HDTV.x264-AVS", "Sleepy Hollow"},
		{"NBC.Nightly.News.2016.02.17.WEB-DL.x264-2Maverick", "NBC Nightly News"},
		{"The.Daily.Show.with.Trevor.Noah.2016.02.08.Gillian.Jacobs.720p.CC.AAC2.0.x264-monkee", "The Daily Show with Trevor Noah"},
		{"The.Daily.Show.with.Trevor  Noah.2016.02.08.Gillian.Jacobs.720p.CC.AAC2.0.x264-monkee", "The Daily Show with Trevor Noah"},
	}

	for _, ex := range expectations {
		res := parseNameToTVName(ex[0])
		if res != ex[1] {
			t.Errorf("Expected %s to parse to %s got %s", ex[0], ex[1], res)
		}
	}
}

func TestClean(t *testing.T) {
	expected := [][]string{
		{"àáâãäæÀÁÂÃÄ'", "aaaaaaAAAAA"},
		{"çÇ", "cC"},
		{"ΣèéêëÈÉÊË", "eeeeeEEEE"},
		{"ìíîïÌÍÎÏ", "iiiiIIII"},
		{"òóôõöÒÓÔÕÖ", "oooooOOOOO"},
		{"ß", "ss"},
		{"ùúûüūÚÛÜŪ", "uuuuuUUUU"},
		{`"History Channel_(Foo:Bar)#'Baz'!" `, "FooBarBaz"},
	}
	for _, ex := range expected {
		res := cleanName(ex[0])
		if res != ex[1] {
			t.Errorf("Expected %s to parse to %s got %s", ex[0], ex[1], res)
		}
	}
}
