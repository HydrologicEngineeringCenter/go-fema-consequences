package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
func Test_NON_AWS_Config_Write(t *testing.T) {
	c := config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "gpkg"}
	bytes, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	w, err := os.OpenFile("/workspaces/go-fema-consequences/data/example.eventconfig", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	w.Write(bytes)
	w.Close()
}
func Test_NON_AWS_Config_Read(t *testing.T) {
	c := config.FromFile("/workspaces/go-fema-consequences/data/example.eventconfig")
	comp, err := compute.Init(c)
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
