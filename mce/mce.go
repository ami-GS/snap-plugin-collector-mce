package mce

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	//"github.com/intelsdi-x/snap/control/plugin"
	//"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	//"github.com/intelsdi-x/snap/core"
)

const (
	vendorName    = "ami-GS"
	PluginName    = "mce"
	PluginVersion = 1
)

var mceLog = "/var/log/mcelog"

const metricAll string = "everything"

// for first testing
var AllMetricsNames []string = []string{
	"cpu",
	"memory",
	"cache",
	"IO",
}

type MCECollector struct {
	// this is used for checking file change
	prevTimeStamp string
	// this is decided by mcelog process argument
	availableMetrics []string
}

func (p *MCECollector) GetMetricTypes(_ plugin.Config) ([]plugin.Metric, error) {
	metricTypes := []plugin.Metric{}
	for i := 0; i < len(p.availableMetrics); i++ {
		metricType := plugin.Metric{
			Namespace: plugin.NewNamespace(vendorName, PluginName, p.availableMetrics[i]),
		}
		metricTypes = append(metricTypes, metricType)
	}
	return metricTypes, nil
}

func (p *MCECollector) CollectMetrics(metricTypes []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	fi, err := os.Stat(mceLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "k%s was not found, did you instll mcelog?\n", mceLog)
		return metrics, nil
	}
	modTime := fi.ModTime().String()

	if p.prevTimeStamp != modTime {
		p.prevTimeStamp = modTime
		ts := time.Now()
		data, err := getParsedData(mceLog)
		if err != nil {
			return nil, err
		}

		for _, metricType := range metricTypes {
			ns := metricType.Namespace
			metric := plugin.Metric{
				Namespace: ns,
				Data:      data,
				Timestamp: ts,
				Version:   PluginVersion,
			}
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}

func (p *MCECollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{vendorName, "???", PluginName}, "key", false, plugin.SetDefaultString(mceLog))
	return *policy, nil
}

// New creates instance of interface info plugin
func New() *MCECollector {
	// TODO : check mcelog process argument, trigger script whether it is avairable metric
	metrics := []string{"cpu", "memory", metricAll}
	return &MCECollector{
		prevTimeStamp:    "",
		availableMetrics: metrics,
	}
}

func getParsedData(path string) (string, error) {
	// TODO : return not string, but []???? for each log
	file, err := os.Open(mceLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, mceLog+" Open \n")
		return "", err
	}
	defer file.Close()

	var data string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		// TODO : parser for each metrics
		data += sc.Text()
	}
	return data, nil
}
