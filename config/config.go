package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

var(
	Conf Config
)

func Load() {
	raw, err := ioutil.ReadFile("config.toml"); if err != nil {
		panic(err)
	}

	if _, err = toml.Decode(string(raw), &Conf); err != nil {
		panic(err)
	}
}
