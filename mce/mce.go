package mce

import (
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

const (
	vendorName    = "ami-GS"
	pluginName    = "mce"
	pluginVersion = 0.1
	pluginType    = plugin.CollectorPluginType
)

var mceLog = "/var/log/mcelog"

// for first testing
const metricAll string = "everything"

type Plugin struct{}

func (p *Plugin) GetMetricTypes(_ plugin.ConfigType) ([]plugin.Metric, error) {
	metricTypes := []plugin.Metric{}
	metricType := plugin.Metric{
		Namespace_: core.NewNamespace(vendorName, pluginName, metricAll),
	}
	metricTypes = append(metricTypes, metricType)

	return metricTypes, nil
}

func (p *Plugin) CollectMetrics(metricTypes []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	for _, metricType := range metricTypes {
		ns := metricType.Namespace

		metric := plugin.Metric{
			Namespace: ns,
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (p *Plugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}

// New creates instance of interface info plugin
func New() *Plugin {
	fh, err := os.Open(mceLog)
	if err != nil {
		// TODO : need to check whether file does not exists or mcelog was not installed
		return nil
	}
	defer fh.Close()

	hwinfo, err := getBasicHWInfo()
	if err != nil {
		return nil
	}

	return &Plugin{}
}

func getStats() {}

func getBasicHWInfo() (*HWinfo, error) {
	cpuNum := 1
	dimmNum := 2
	return &HWinfo{
		cpuNum:  cpuNum,
		dimmNum: dimmNum,
	}, nil

}

type HWinfo struct {
	cpuNum  int64
	dimmNum int64
}
