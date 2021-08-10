package compute

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/HydrologicEngineeringCenter/go-fema-consequences/outputwriter"
	consequences_compute "github.com/USACE/go-consequences/compute"
	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazardproviders"
	"github.com/USACE/go-consequences/structureprovider"
)

type Compute struct {
	Hpfp           string
	Shp_FP         string
	Ow             consequences.ResultsWriter
	TempFileOutput string
}

func Init(fp string, sfp string) (Compute, error) {
	var err error
	err = nil
	var hp hazardproviders.HazardProvider //not sure what it will be yet, but we can declare it!
	var se error
	se = nil
	//grab the tif file key, change the directory to inventory/ORNLcentroids_LBattributes.shp
	//parts := strings.Split(fp, "/")

	//sfp := strings.Replace(fp, parts[len(parts)-1], "inventory/ORNLcentroids_LBattributes.shp", -1)

	var he error
	hp, he = hazardproviders.Init(fp)

	ofp := fp
	// pull the .tif off the end?
	var oe error
	oe = nil
	if len(ofp) > 4 {
		ofp = ofp[:len(ofp)-4] //good enough for government work?
		//write to temp directory and copy then paste!
		ofp = "/app/working/" + filepath.Base(ofp)
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
	if he == nil {
		hp.Close()
	}

	return Compute{Hpfp: fp, Shp_FP: sfp, TempFileOutput: ofp}, err
}
func (c Compute) Compute_NSI() {
	ofp := c.TempFileOutput
	nsisp := structureprovider.InitNSISP()
	now, err := consequences.InitGpkResultsWriter(ofp+"_consequences_nsi.gpkg", "results")
	if err != nil {
		log.Println(err)
	}
	nows, err := consequences.InitShpResultsWriter(ofp+"_consequences_nsi.shp", "results")
	if err != nil {
		log.Println(err)
	}
	nowgs, err := consequences.InitGeoJsonResultsWriterFromFile(ofp + "_consequences_nsi.json")
	if err != nil {
		log.Println(err)
	}
	nowsdollars, err := consequences.InitSummaryResultsWriterFromFile(ofp + "_summaryDollars_nsi.csv")
	if err != nil {
		log.Println(err)
	}
	nowsdepths := outputwriter.InitSummaryByDepth(ofp + "_summaryDepths_nsi.csv")

	compute(c.Hpfp, nsisp, now)
	compute(c.Hpfp, nsisp, nows)
	compute(c.Hpfp, nsisp, nowgs)
	compute(c.Hpfp, nsisp, nowsdollars)
	compute(c.Hpfp, nsisp, nowsdepths)
}
func (c Compute) Compute_SHP() error {
	ofp := c.TempFileOutput
	sp, err := structureprovider.InitSHP(c.Shp_FP)
	if err != nil {
		log.Println(err)
		return err
	}
	ow, err := consequences.InitGpkResultsWriter(ofp+"_consequences.gpkg", "results")
	if err != nil {
		log.Println(err)
		return err
	}
	ows, err := consequences.InitShpResultsWriter(ofp+"_consequences.shp", "results")
	if err != nil {
		log.Println(err)
		return err
	}
	owgs, err := consequences.InitGeoJsonResultsWriterFromFile(ofp + "_consequences.json")
	if err != nil {
		log.Println(err)
		return err
	}
	owsdollars, err := consequences.InitSummaryResultsWriterFromFile(ofp + "_summaryDollars.csv")
	if err != nil {
		log.Println(err)
		return err
	}
	owsdepths := outputwriter.InitSummaryByDepth(ofp + "_summaryDepths.csv")

	compute(c.Hpfp, sp, ow)
	compute(c.Hpfp, sp, ows)
	compute(c.Hpfp, sp, owgs)
	compute(c.Hpfp, sp, owsdollars)
	compute(c.Hpfp, sp, owsdepths)
	return nil
}
func compute(hpfp string, sp consequences.StreamProvider, ow consequences.ResultsWriter) {
	hp, err := hazardproviders.Init(hpfp)
	if err != nil {
		log.Println(err)
	}
	defer hp.Close()
	defer ow.Close()
	consequences_compute.StreamAbstract(hp, sp, ow)
}
