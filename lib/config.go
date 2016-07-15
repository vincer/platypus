package lib

type ConfigType struct {
	CacheTimeoutSeconds int
	Ip                  string
	Port                int
	Verbose             bool
}

// MAIN CONFIG
var Config ConfigType
