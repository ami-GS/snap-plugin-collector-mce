package main

import (
	"github.com/ami-GS/snap-plugin-collector-mce/mce_ondemand/mce"
	"github.com/ami-GS/snap-plugin-collector-mce/mce_stream/mce_stream"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

func main() {
	plugin.StartStreamCollector(mceStream.NewStream(mce.MceLogPath), mce.PluginName, mce.PluginVersion)
}
