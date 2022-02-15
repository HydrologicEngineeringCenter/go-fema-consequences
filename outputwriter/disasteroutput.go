package outputwriter
import (
	"fmt"
	"io"
	"os"

	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazards"
)

type disasterOuput struct {
	filepath   string
	w          io.Writer
	fipsmap map[string]countyRecord
}
type countyRecord struct{
	statefips string
	countyfips string
	resDamCount int
	nonresdamcount int
	resTotDam float64
	nonResTotDam float64
	byAssetType map[string]typeRecord
}
type typeRecord struct{
	totalInCounty int
	destroyed int
	majorDamage int
	minorDamage int
	affected int
	totalDamages float64
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
func (srw *summaryByDepth) Write(r consequences.Result) {
	fips := r.Fetch("cbfips")
	ssccc := fips[0:5]//first five ditgits of fips are state (2) and county (3)
	cr, crok := srw.fipsmap[ssccc]
	if crok{
		//previously existed, update and reassign.
		cr.Update(r)
		srw.fipsmap[ssccc] = cr
	}else{
		am := make(map[string]typeRecord)
		newcr := countyRecord{statefips: fips[0:2], countyfips: fips[2:5], byAssetType: am}
		newcr.Update(r)
		srw.fipsmap[ssccc] = newcr
	}
}
func (cr *countyRecord) Update(r consequences.Result){
	damcat := r.Fetch("damage category")
	if damcat == "res"{
		cr.resDamCount +=1
		cr.resTotDam += r.Fetch("structure damage").(float64)
		cr.resTotDam += r.Fetch("content damage").(float64)
	}else{
		cr.nonresdamcount +=1
		cr.nonResTotDam += r.Fetch("structure damage").(float64)
		cr.nonResTotDam += r.Fetch("content damage").(float64)
	}
	//add to asset type map.
	/*
	Residential
	Agriculture
	Assembly
	Commercial
	Education
	Government
	Industrial
	Non-Profit
	Unclassified
	Utility and Miscellaneous
	*/
}
func (srw *summaryByDepth) Close() {
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