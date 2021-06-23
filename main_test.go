package main

import (
	"bytes"
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

func createConfigs_nonAWS() []config.Config {
	cs := make([]config.Config, 5)
	cs[0] = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "gpkg"}
	cs[1] = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "nhc", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "shp"}
	cs[2] = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "geojson"}
	cs[3] = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "summaryDollars"}
	cs[4] = config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "nhc", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "summaryDepths"}
	return cs
}
func createConfigs_AWS() []config.Config {
	cs := make([]config.Config, 5)
	cs[0] = config.Config{Hfp: "/vsis3/media/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/vsis3/media/nsi.gpkg", Ss: "gpkg", Ot: "gpkg", Ofp: "/results"}
	cs[1] = config.Config{Hfp: "/vsis3/media/clipped_sample.tif", HpSource: "nhc", HpUnits: "feet", Sfp: "/vsis3/media/nsi.gpkg", Ss: "gpkg", Ot: "shp", Ofp: "/results"}
	cs[2] = config.Config{Hfp: "/vsis3/media/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/vsis3/media/nsi.gpkg", Ss: "gpkg", Ot: "geojson", Ofp: "/results"}
	cs[3] = config.Config{Hfp: "/vsis3/media/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/vsis3/media/nsi.gpkg", Ss: "gpkg", Ot: "summaryDollars", Ofp: "/results"}
	cs[4] = config.Config{Hfp: "/vsis3/media/clipped_sample.tif", HpSource: "nhc", HpUnits: "feet", Sfp: "/vsis3/media/nsi.gpkg", Ss: "gpkg", Ot: "summaryDepths", Ofp: "/results"}
	return cs
}
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
	c := config.Config{Hfp: "/vsis3/media/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/vsis3/media/nsi.gpkg", Ss: "gpkg", Ot: "gpkg", Ofp: "/results"} //config.Config{Hfp: "/workspaces/go-fema-consequences/data/clipped_sample.tif", HpSource: "depths", HpUnits: "feet", Sfp: "/workspaces/go-fema-consequences/data/nsi.gpkg", Ss: "gpkg", Ot: "gpkg"}
	bytes, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	w, err := os.OpenFile("/workspaces/go-fema-consequences/data/new.eventconfig", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
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
func Test_Consequences_Events(t *testing.T) {
	response, err := http.Get("http://host.docker.internal:8000/fema-consequences/events/")
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
func Test_Compute(t *testing.T) {
	configs := createConfigs_AWS()
	for _, c := range configs {
		b, _ := json.Marshal(c)
		response, err := http.Post(
			"http://host.docker.internal:8000/fema-consequences/compute",
			"application/json; charset=UTF-8",
			bytes.NewReader(b),
		)
		if err != nil {
			log.Fatal(err)
		}

		defer response.Body.Close()
		result, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s", result)
	}

}
