package main

import conf "github.com/eth-library/dataset-dj/configuration"

var (
	config *conf.HandlerConfig
)

func setupConfig() {
	config = conf.InitHandlerConfig()
}

func main() {
	setupConfig()
	clientLoop()
}
