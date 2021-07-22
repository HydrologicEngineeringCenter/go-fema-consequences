package compute

import (
	"errors"
	"path/filepath"
	"strings"

	consequences_compute "github.com/USACE/go-consequences/compute"
	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazardproviders"
	"github.com/USACE/go-consequences/structureprovider"
)

type Compute struct {
	Hp                   hazardproviders.HazardProvider
	NSI_Sp               consequences.StreamProvider
	Shp_Sp               consequences.StreamProvider
	Ow                   consequences.ResultsWriter
	TempFileOutput       string
	NSI_OutputFolderPath string
	SHP_OutputFolderPath string
}

func Init(fp string, outputdir string) (Compute, error) {
	var err error
	err = nil
	var sp consequences.StreamProvider
	var hp hazardproviders.HazardProvider //not sure what it will be yet, but we can declare it!
	var ow consequences.ResultsWriter     //need a file path to write anything...
	var se error
	se = nil
	//grab the tif file key, change the directory to inventory/ORNLcentroids_LBattributes.shp
	parts := strings.Split(fp, "/")

	sfp := strings.Replace(fp, parts[len(parts)-1], "inventory/ORNLcentroids_LBattributes.shp", -1)
	//add /vsis3/?
	sp, se = structureprovider.InitSHP(sfp)

	var he error
	hp, he = hazardproviders.Init(fp)

	ofp := fp
	// pull the .tif off the end?
	var oe error
	oe = nil
	if len(ofp) > 4 {
		ofp = ofp[:len(ofp)-4] //good enough for government work?
		// pull vsis3 off the front!
		//write to temp directory and copy then paste!
		ofp = "/app/working/" + filepath.Base(ofp)
		ofp += "_consequences.gpkg"
		ow, oe = consequences.InitGpkResultsWriter(ofp, "results")
		/*
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
		*/
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
	nsisp := structureprovider.InitNSISP()
	return Compute{Hp: hp, NSI_Sp: nsisp, Shp_Sp: sp, Ow: ow, NSI_OutputFolderPath: "NSI_" + outputdir, SHP_OutputFolderPath: "SHP_" + outputdir, TempFileOutput: ofp}, err
}
func (c Compute) Compute() {
	compute(c.Hp, c.NSI_Sp, c.Ow)
	//compute(c.Hp, c.Shp_Sp, c.Ow)
}
func compute(hp hazardproviders.HazardProvider, sp consequences.StreamProvider, ow consequences.ResultsWriter) {
	defer hp.Close()
	defer ow.Close()
	consequences_compute.StreamAbstract(hp, sp, ow)
}
