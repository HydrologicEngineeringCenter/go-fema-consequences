package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/HydrologicEngineeringCenter/go-fema-consequences/compute"
)

func Test_NON_AWS_Compute(t *testing.T) {
	fp := "/workspaces/go-fema-consequences/data/clipped_sample.tif"
	comp, err := compute.Init(fp, "")
	if err != nil {
		panic(err)
	}
	comp.Compute_NSI()
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
	response, err := http.Get("http://host.docker.internal:8000/fema-consequences/events")
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
func Test_IsLive_CWBI(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	response, err := client.Get("https://ml-dev.sec.usace.army.mil/nsi-ml/fema-consequences")
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
