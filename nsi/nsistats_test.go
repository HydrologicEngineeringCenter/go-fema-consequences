package nsi

import (
	"fmt"
	"testing"

	"github.com/USACE/go-consequences/structureprovider"
)

func TestNsiStatsByFips(t *testing.T) {
	var fips string = "15005" //Kalawao county (smallest county in the us by population)stats?bbox=-81.58418,30.25165,-81.58161,30.26939,-81.55898,30.26939,-81.55281,30.24998,-81.58418,30.25165
	nsisp := structureprovider.InitNSISP()
	stats := StatsByFips(fips, nsisp)
	fmt.Println(stats)
	if stats.TotalCount != 101 {
		t.Errorf("GetByFips(%s) yeilded %v structures; expected 101", fips, stats.TotalCount)
	}
}
