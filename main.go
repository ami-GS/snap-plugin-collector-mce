package main

import (
	"github.com/ami-GS/snap-plugin-collector-mce/mce"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

func main() {
	plugin.StartCollector(mce.New(), mce.PluginName, mce.PluginVersion, plugin.ConcurrencyCount(1))

}
