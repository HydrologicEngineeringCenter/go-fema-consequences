package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Config struct {
	Hfp      string `json:"hfp"`
	HpSource string `json:"hs"`
	HpUnits  string `json:"ht"`
	Sfp      string `json:"sfp"`
	Ss       string `json:"ss"`
	Ofp      string `json:"ofp"`
	Ot       string `json:"ot"`
}

func FromFile(fp string) (Config, error) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		panic(err)
	}
	c := Config{}
	errm := json.Unmarshal(b, &c)

	return c, errm
}
func (c Config) Validate() error {
	var s string
	s = ""
	haserrors := false
	if c.Hfp == "" {
		s += "Hazard File Path is Empty"
		haserrors = true
	}
	if c.Ss != "nsi" {
		if c.Sfp == "" {
			s += "Structure Source is not set to NSI and Structure File Path is Empty"
			haserrors = true
		}
	}
	if haserrors {
		return errors.New(s)
	}
	return nil
}
