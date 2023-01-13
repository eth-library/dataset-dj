package main

import conf "github.com/eth-library/dataset-dj/configuration"

var (
	config         *conf.HandlerConfig
	libDriveConfig *LibDriveConfig
	secrets        *Secrets
)

func setupConfig() {
	config = conf.InitHandlerConfig()
	libDriveConfig = loadLibDriveConfig("./conf.d/")
	secrets = loadSecrets("./conf.d/")
}

func main() {
	setupConfig()
	clientLoop()
}
