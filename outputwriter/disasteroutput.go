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
	
	return &disasterOuput{filepath: filepath, w: w}
}
func (srw *summaryByDepth) Write(r consequences.Result) {

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