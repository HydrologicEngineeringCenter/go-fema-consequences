package compute

import (
	"github.com/HydrologicEngineeringCenter/go-fema-consequences/config"
	consequences_compute "github.com/USACE/go-consequences/compute"
	"github.com/USACE/go-consequences/consequences"
	"github.com/USACE/go-consequences/hazardproviders"
)

type compute struct {
	Hp hazardproviders.HazardProvider
	Sp consequences.StreamProvider
	Ow consequences.ResultsWriter
}

func Init(c config.Config) compute {
	return compute{}
}
func (c compute) Compute() {
	consequences_compute.StreamAbstract(c.Hp, c.Sp, c.Ow)
}
