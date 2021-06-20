package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/HydrologicEngineeringCenter/go-fema-consequences/compute"
	"github.com/HydrologicEngineeringCenter/go-fema-consequences/config"
)

func Test_NON_AWS_Config_To_Compute(t *testing.T) {
	c := config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "gpkg"}
	comp, err := compute.Init(c)
	if err != nil {
		panic(err)
	}
	comp.Compute()
	c = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "nhc", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "shp"}
	comp, err = compute.Init(c)
	if err != nil {
		panic(err)
	}
	comp.Compute()
	c = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "geojson"}
	comp, err = compute.Init(c)
	if err != nil {
		panic(err)
	}
	comp.Compute()
	c = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "summaryDollars"}
	comp, err = compute.Init(c)
	if err != nil {
		panic(err)
	}
	comp.Compute()
	c = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "nhc", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "summaryDepths"}
	comp, err = compute.Init(c)
	if err != nil {
		panic(err)
	}
	comp.Compute()
}
func Test_Consequences_IsLive(t *testing.T) {
	response, err := http.Get("http://host.docker.internal:8000/fema-consequences")
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", b)
}
