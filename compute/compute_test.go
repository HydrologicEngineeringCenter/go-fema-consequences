package compute

import (
	"testing"

	"github.com/HydrologicEngineeringCenter/go-fema-consequences/outputwriter"
	"github.com/USACE/go-consequences/structureprovider"
)

func TestComputeWithEcam(t *testing.T) {
	hfp := "/workspaces/go-fema-consequences/data/ida/Coastal_Louisiana_28Aug2021_WGS84.tif"
	outputfp := "/workspaces/go-fema-consequences/data/ida/Coastal_Louisiana_28Aug2021_WGS84_results.csv"
	sp := structureprovider.InitNSISP()
	ow := outputwriter.InitDisasterOutput(outputfp, sp)
	ComputeResults(hfp, sp, ow)
}
