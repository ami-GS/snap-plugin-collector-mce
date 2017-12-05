package main

import (
	"flag"

	"github.com/ami-GS/snap-plugin-collector-mce/mce"
	"github.com/ami-GS/snap-plugin-collector-mce/mce_stream"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

func main() {
	// Currently user defined flag is not available
	var streaming bool
	flag.BoolVar(&streaming, "stream", false, "streaming flag")
	flag.BoolVar(&streaming, "s", false, "streaming flag")
	flag.Parse()

	// TODO : should be configurable?
	//if streaming {
	plugin.StartStreamCollector(mceStream.NewStream(mce.MceLogPath), mce.PluginName, mce.PluginVersion)
	//} else {
	//plugin.StartCollector(mce.New(mce.MceLogPath), mce.PluginName, mce.PluginVersion, plugin.ConcurrencyCount(1))
	//}
}
