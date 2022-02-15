package outputwriter

import (
	"io"
	"os"

	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazards"
)

type disasterOuput struct {
	filepath string
	w        io.Writer
	fipsmap  map[string]countyRecord
}
type countyRecord struct {
	statefips      string
	countyfips     string
	resDamCount    int
	nonresdamcount int
	resTotDam      float64
	nonResTotDam   float64
	byAssetType    map[string]typeRecord
}
type typeRecord struct {
	totalInCounty        int
	damageCategorization map[string]int
	thresholds           map[string]float64
	totalDamages         float64
}

func InitDisasterOutput(filepath string) *disasterOuput {
	w, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	//make fipsmap
	fm := make(map[string]countyRecord)
	return &disasterOuput{filepath: filepath, w: w, fipsmap: fm}
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
			v, _ = r.Fetch("content damage")
			damage = v.(float64)
			cr.resTotDam += damage
		} else {
			cr.nonresdamcount += 1
			v, _ := r.Fetch("structure damage") //unsafe skipping error.
			damage := v.(float64)
			cr.nonResTotDam += damage
			v, _ = r.Fetch("content damage")
			damage = v.(float64)
			cr.nonResTotDam += damage
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
	/*	fmt.Fprintf(srw.w, "Outcome, Count\n")
		h := srw.totals
		for i, v := range h {
			fmt.Fprintf(srw.w, fmt.Sprintf("%v, %v\n", srw.headers[i], v))
		}
		fmt.Fprintf(srw.w, fmt.Sprintf("Total Building Count, %v\n", srw.grandTotal))
		w2, ok := srw.w.(io.WriteCloser)
		if ok {
			w2.Close()
		}
	*/
}
