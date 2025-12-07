package config

type InstrumentProfile struct {
	Name    string
	MinFreq float64
	MaxFreq float64
}

var Profiles = map[string]InstrumentProfile{
	"guitar": {
		Name:    "Guitar (Standard)",
		MinFreq: 75.0,   // Slightly below E2
		MaxFreq: 1400.0, // Slightly above E6
	},
	"bouzouki": {
		Name:    "Bouzouki (6 or 8 string)",
		MinFreq: 110.0, // Slightly below C3
		MaxFreq: 360.0, // Slightly above F4
	},
}

func GetProfile(name string) (InstrumentProfile, bool) {
	profile, ok := Profiles[name]
	return profile, ok
}

func ListInstruments() []string {
	names := make([]string, 0, len(Profiles))
	for name := range Profiles {
		names = append(names, name)
	}

	return names
}
