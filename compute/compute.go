package compute

import (
	"errors"
	"path/filepath"

	"github.com/HydrologicEngineeringCenter/go-fema-consequences/config"
	"github.com/HydrologicEngineeringCenter/go-fema-consequences/outputwriter"
	"github.com/HydrologicEngineeringCenter/go-tc-consequences/nhc"
	consequences_compute "github.com/USACE/go-consequences/compute"
	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazardproviders"
	"github.com/USACE/go-consequences/structureprovider"
)

type Compute struct {
	Hp               hazardproviders.HazardProvider
	Sp               consequences.StreamProvider
	Ow               consequences.ResultsWriter
	TempFileOutput   string
	OutputFolderPath string
}

func Init(c config.Config) (Compute, error) {
	var err error
	err = nil
	var sp consequences.StreamProvider
	var hp hazardproviders.HazardProvider //not sure what it will be yet, but we can declare it!
	hazardProviderVerticalUnitsIsFeet := true
	var ow consequences.ResultsWriter //need a file path to write anything...
	var se error
	se = nil
	if c.Sfp != "" {
		switch c.Ss {
		case "gpkg":
			sp, se = structureprovider.InitGPK(c.Sfp, "nsi")
		case "shp":
			sp, se = structureprovider.InitSHP(c.Sfp)
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
		se = errors.New("cannot compute without hazard provider path")
	}
	var he error
	he = nil
	if c.Hfp != "" {
		switch c.HpSource {
		case "nhc":
			hp = nhc.Init(c.Hfp)
		case "depths":
			if hazardProviderVerticalUnitsIsFeet {
				hp, he = hazardproviders.Init(c.Hfp)
			} else {
				hp, he = hazardproviders.Init_Meters(c.Hfp)
			}
		default: //assume depth
			if hazardProviderVerticalUnitsIsFeet {
				hp, he = hazardproviders.Init(c.Hfp)
			} else {
				hp, he = hazardproviders.Init_Meters(c.Hfp)
			}
		}
	} else {
		he = errors.New("cannot compute without hazard provider path")
	}
	ofp := c.Hfp
	// pull the .tif off the end?
	var oe error
	oe = nil
	if len(ofp) > 4 {
		ofp = ofp[:len(ofp)-4] //good enough for government work?
		// pull vsis3 off the front!
		//write to temp directory and copy then paste!
		ofp = "/app/working/" + filepath.Base(ofp)

		if ofp != "" {
			switch c.Ot {
			case "gpkg":
				ofp += "_consequences.gpkg"
				ow, oe = consequences.InitGpkResultsWriter(ofp, "results")
			case "shp":
				ofp += "_consequences.shp"
				ow, oe = consequences.InitShpResultsWriter(ofp, "results")
			case "geojson":
				ofp += "_consequences.json"
				ow, oe = consequences.InitGeoJsonResultsWriterFromFile(ofp)
			case "summaryDollars":
				ofp += "_summaryDollars.csv"
				ow, oe = consequences.InitSummaryResultsWriterFromFile(ofp)
			case "summaryDepths":
				ofp += "_summaryDepths.csv"
				ow = outputwriter.InitSummaryByDepth(ofp)
			default:
				ofp += "_consequences.gpkg"
				ow, oe = consequences.InitGpkResultsWriter(ofp, "results")
			}
		} else {
			oe = errors.New("we need an input hazard file path use so we can define the output path")
		}
	} else {
		oe = errors.New("Output file is shorter than 4 characters, which seems odd... " + ofp)
	}

	//consolidate errors to one error message.
	if se != nil {
		if he != nil {
			if oe != nil {
				err = errors.New(se.Error() + "\n" + he.Error() + "\n" + oe.Error() + "\n")
			} else {
				err = errors.New(se.Error() + "\n" + he.Error() + "\n")
			}
		} else {
			err = errors.New(se.Error() + "\n")
		}
	} else if he != nil {
		if oe != nil {
			err = errors.New(he.Error() + "\n" + oe.Error() + "\n")
		} else {
			err = errors.New(he.Error() + "\n")
		}
	} else {
		if oe != nil {
			err = errors.New(oe.Error() + "\n")
		}
	}

	return Compute{Hp: hp, Sp: sp, Ow: ow, OutputFolderPath: c.Ofp, TempFileOutput: ofp}, err
}
func (c Compute) Compute() {
	defer c.Hp.Close()
	defer c.Ow.Close()
	consequences_compute.StreamAbstract(c.Hp, c.Sp, c.Ow)
}
