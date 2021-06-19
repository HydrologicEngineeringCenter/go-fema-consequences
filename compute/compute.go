package compute

import (
	"errors"

	"github.com/HydrologicEngineeringCenter/go-fema-consequences/config"
	"github.com/HydrologicEngineeringCenter/go-fema-consequences/outputwriter"
	"github.com/HydrologicEngineeringCenter/go-tc-consequences/nhc"
	consequences_compute "github.com/USACE/go-consequences/compute"
	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazardproviders"
	"github.com/USACE/go-consequences/structureprovider"
)

type compute struct {
	Hp hazardproviders.HazardProvider
	Sp consequences.StreamProvider
	Ow consequences.ResultsWriter
}

func Init(c config.Config) (compute, error) {
	var err error
	var sp consequences.StreamProvider
	var hp hazardproviders.HazardProvider //not sure what it will be yet, but we can declare it!
	hazardProviderVerticalUnitsIsFeet := true
	var ow consequences.ResultsWriter //need a file path to write anything...
	if c.Sfp != "" {
		switch c.Ss {
		case "gpkg":
			sp = structureprovider.InitGPK(c.Sfp, "nsi")
		case "shp":
			sp = structureprovider.InitSHP(c.Sfp)
		case "nsi":
			sp = structureprovider.InitNSISP() //default to NSI API structure provider.
		default:
			sp = structureprovider.InitNSISP()
		}
	} else {
		sp = structureprovider.InitNSISP()
	}
	if c.HpUnits != "" {
		switch c.HpUnits {
		case "feet":
			hazardProviderVerticalUnitsIsFeet = true
		case "meters":
			hazardProviderVerticalUnitsIsFeet = false
		}
	} else {
		err = errors.New("cannot compute without hazard provider path")
	}
	if c.Hfp != "" {
		switch c.HpSource {
		case "nhc":
			hp = nhc.Init(c.Hfp)
		case "depths":
			if hazardProviderVerticalUnitsIsFeet {
				hp = hazardproviders.Init(c.Hfp)
			} else {
				hp = hazardproviders.Init_Meters(c.Hfp)
			}
		}
	} else {
		err = errors.New("cannot compute without hazard provider path")
	}
	ofp := c.Hfp
	// pull the .tif off the end?
	ofp = ofp[:len(ofp)-4] //good enough for government work?
	if ofp != "" {
		switch c.Ot {
		case "gpkg":
			ofp += "_consequences.gpkg"
			ow = consequences.InitGpkResultsWriter(ofp, "results")
		case "shp":
			ofp += "_consequences.shp"
			ow = consequences.InitShpResultsWriter(ofp, "results")
		case "geojson":
			ofp += "_consequences.json"
			ow = consequences.InitGeoJsonResultsWriterFromFile(ofp)
		case "summaryDollars":
			ofp += "_summaryDollars.csv"
			ow = consequences.InitSummaryResultsWriterFromFile(ofp)
		case "summaryDepths":
			ofp += "_summaryDepths.csv"
			ow = outputwriter.InitSummaryByDepth(ofp)
		default:
			ofp += "_consequences.gpkg"
			ow = consequences.InitGpkResultsWriter(ofp, "results")
		}
	} else {
		err = errors.New("we need an input hazard file path use")
	}
	return compute{Hp: hp, Sp: sp, Ow: ow}, err
}
func (c compute) Compute() {
	defer c.Hp.Close()
	defer c.Ow.Close()
	consequences_compute.StreamAbstract(c.Hp, c.Sp, c.Ow)
}
