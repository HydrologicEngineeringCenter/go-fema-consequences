package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Hfp      string `json:"hfp"`
	HpSource string `json:"hs"`
	HpUnits  string `json:"ht"`
	Sfp      string `json:"sfp"`
	Ss       string `json:"ss"`
	Ot       string `json:"ot"`
}

func FromFile(fp string) Config {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		panic(err)
	}
	c := Config{}
	json.Unmarshal(b, &c)
	return c
}
