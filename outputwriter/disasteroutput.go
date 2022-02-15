package outputwriter

import (
	"io"
	"os"

	"github.com/HydrologicEngineeringCenter/go-fema-consequences/nsi"
	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazards"
	"github.com/USACE/go-consequences/indirecteconomics"
	"github.com/USACE/go-consequences/structureprovider"
)

type disasterOuput struct {
	filepath string
	w        io.Writer
	fipsmap  map[string]countyRecord
	sp       structureprovider.StructureProvider
}
type countyRecord struct {
	statefips      string
	countyfips     string
	resDamCount    int
	nonresdamcount int
	resTotDam      float64
	indirectLosses float64 //in millions.
	workingPop     int32
	nonResTotDam   float64
	totalValue     float64
	totalDamages   float64
	byAssetType    map[string]typeRecord
}
type typeRecord struct {
	totalInCounty        int
	damageCategorization map[string]int
	thresholds           map[string]float64
	totalDamages         float64
}

func InitDisasterOutput(filepath string, sp structureprovider.StructureProvider) *disasterOuput {
	w, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	//make fipsmap
	fm := make(map[string]countyRecord)
	return &disasterOuput{filepath: filepath, w: w, fipsmap: fm, sp: sp}
}
func (srw *disasterOuput) Write(r consequences.Result) {
	f, ferr := r.Fetch("cbfips")

	if ferr == nil {
		fips := f.(string)
		ssccc := fips[0:5] //first five ditgits of fips are state (2) and county (3)
		cr, crok := srw.fipsmap[ssccc]
		if crok {
			//previously existed, update and reassign.
			cr.Update(r)
			srw.fipsmap[ssccc] = cr
		} else {
			am := make(map[string]typeRecord)
			cr = countyRecord{statefips: fips[0:2], countyfips: fips[2:5], byAssetType: am}
			cr.Update(r)
			srw.fipsmap[ssccc] = cr
		}
		srw.fipsmap[ssccc] = cr
	}
}
func (cr *countyRecord) Update(r consequences.Result) {
	d, derr := r.Fetch("damage category")
	if derr == nil {
		damcat := d.(string)
		if damcat == "res" {
			cr.resDamCount += 1
			v, _ := r.Fetch("structure damage") //unsafe skipping error.
			damage := v.(float64)
			cr.resTotDam += damage
			cr.totalDamages += damage
			v, _ = r.Fetch("content damage")
			damage = v.(float64)
			cr.resTotDam += damage
			cr.totalDamages += damage
			pop, _ := r.Fetch("pop2amu65")
			cr.workingPop += pop.(int32)
		} else {
			cr.nonresdamcount += 1
			v, _ := r.Fetch("structure damage") //unsafe skipping error.
			damage := v.(float64)
			cr.nonResTotDam += damage
			cr.totalDamages += damage
			v, _ = r.Fetch("content damage")
			damage = v.(float64)
			cr.nonResTotDam += damage
			cr.totalDamages += damage
		}
	}
	o, oerr := r.Fetch("occupancy type")
	assetType := ""
	if oerr == nil {
		occtype := o.(string)
		switch occtype {
		case "REL1":
			assetType = "Assembly"
		case "AGR1":
			assetType = "Agriculture"
		case "EDU1", "EDU2":
			assetType = "Education"
		case "GOV1", "GOV2", "COM6": //why are hospitals encoded to governement?
			assetType = "Government"
		case "IND1", "IND2", "IND3", "IND4", "IND5", "IND6":
			assetType = "Industrial"
		case "COM1", "COM2", "COM3", "COM4", "COM5", "COM7", "COM8", "COM9", "COM10":
			assetType = "Commercial"
		default:
			assetType = "Residential"
		}
		at, atok := cr.byAssetType[assetType]
		if atok {
			at.Update(r)
		} else {
			//create a new one.
			dc := make(map[string]int)
			th := make(map[string]float64)
			th["No Damage (0 ft)"] = 0.0
			th["Affected (<=2 ft)"] = 2.0
			th["Minor Damage (2 - 4 ft)"] = 4.0
			th["Major Damage (4 - 6 ft)"] = 6.0
			th["Destroyed (6+ ft)"] = 9998.0
			at := typeRecord{damageCategorization: dc}
			at.Update(r)
		}
		cr.byAssetType[assetType] = at
	}

}
func (at *typeRecord) Update(r consequences.Result) {
	h, hok := r.Fetch("hazard")
	if hok == nil {
		he := h.(hazards.DepthEvent)
		depth := he.Depth()
		for k, v := range at.thresholds {
			if depth <= v {
				count, cok := at.damageCategorization[k]
				if cok {
					count += 1
				} else {
					count = 1
				}
				at.damageCategorization[k] = count
				break
			}
		}
		v, _ := r.Fetch("structure damage") //unsafe skipping error.
		damage := v.(float64)
		at.totalDamages += damage
		v, _ = r.Fetch("content damage")
		damage = v.(float64)
		at.totalDamages += damage
	}

}
func (srw *disasterOuput) Close() {
	for k, v := range srw.fipsmap {
		s := nsi.StatsByFips(k, srw.sp)
		v.totalValue = s.TotalValue
		for ak, av := range v.byAssetType {
			av.totalInCounty = s.CountByCategory[ak]
		}
		laborLossRatio := float64(v.workingPop) / float64(s.WorkingResPop2AM)
		capitalLossRatio := float64(v.nonResTotDam) / float64(s.NonResidentialValue)
		//compute ecam.
		er, err := indirecteconomics.ComputeEcam(v.statefips, v.countyfips, capitalLossRatio, laborLossRatio)
		if err == nil {
			for _, pr := range er.ProductionImpacts {
				if pr.Sector == "TOTAL" {
					v.indirectLosses = pr.Change
					break
				}
			}
		}
	}
	//write out results.

}
