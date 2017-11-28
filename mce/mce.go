package mce

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	VendorName    = "ami-GS"
	PluginName    = "mce"
	PluginVersion = 1
)

// default logPath
var MceLogPath = "/var/log/mcelog"

const MetricAll string = "everything"

// AllMetricsNames : for first testing
var AllMetricsNames = []string{
	// TODO : enable CPU/0, CPU/1, ADDR/0x1234, like metrics
	"CPU",
	"ADDR",
	"BANK",
	"THERMAL",
	"Corrected",
	// for user defined search
	"Includes",
}

type MCECollector struct {
	// this is used for checking file change
	prevFileTimeStamp string
	// this is used for picking up latest log
	prevLogTimeStamp uint32
	// this is decided by mcelog process argument
	availableMetrics []string
	// this would be mceLog
	LogPath     string
	initialized bool
}

func (p *MCECollector) GetMetricTypes(_ plugin.Config) ([]plugin.Metric, error) {
	metricTypes := []plugin.Metric{}
	//for i := 0; i < len(p.availableMetrics); i++ {
	for _, val := range p.availableMetrics {
		if val == "Includes" {
			for _, valIn := range p.availableMetrics {
				//if val != "Includes" || val != MetricAll {
				if valIn != "Includes" && valIn != MetricAll {
					metricType := plugin.Metric{
						Namespace: plugin.NewNamespace(VendorName, PluginName, "Includes", valIn),
					}
					metricTypes = append(metricTypes, metricType)
				}
			}
		} else {
			metricType := plugin.Metric{
				Namespace: plugin.NewNamespace(VendorName, PluginName, val),
			}
			metricTypes = append(metricTypes, metricType)
		}
	}
	return metricTypes, nil
}

func (p *MCECollector) CollectMetrics(metricTypes []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	if !p.initialized {
		dat, ok := metricTypes[0].Config["logpath"]
		if ok {
			p.LogPath = dat.(string)
		}
		p.initialized = true
	}

	ok, err := p.WasFileUpdated()
	if err != nil {
		return nil, err
	}
	if ok {
		mceLogs, err := p.GetMceLog()
		if err != nil {
			return nil, err
		}
		if len(mceLogs) == 0 {
			return metrics, nil
		}
		metrics = StuffLogToMetrics(mceLogs, metricTypes)
	}
	return metrics, nil
}

func StuffLogToMetrics(mceLogs []MceLogInfo, metricIn []plugin.Metric) (metricOut []plugin.Metric) {
	ts := time.Now()
	for _, metricType := range metricIn {
		ns := metricType.Namespace
		// TODO : smarter method.
		data := ""
		for _, log := range mceLogs {
			key := ns[len(ns)-1].Value
			if ns[len(ns)-2].Value == "Includes" || key == MetricAll {
				data += log.AsItIs + "\n"
				continue
			}
			val, ok := log.data[key]
			if ok {
				data += val
			}
		}
		metric := plugin.Metric{
			Namespace: ns,
			Data:      data, // TODO : use appropriate telemetry
			Timestamp: ts,
			Version:   PluginVersion,
		}
		metricOut = append(metricOut, metric)
	}
	return metricOut
}

func (p *MCECollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	err := policy.AddNewStringRule([]string{VendorName, "var/log", PluginName}, "key", false, plugin.SetDefaultString(p.LogPath))
	if err != nil {
		return *policy, err
	}
	return *policy, nil
}

// New creates instance of interface info plugin
func New(logPath string) *MCECollector {
	metrics := append(AllMetricsNames, MetricAll)
	return &MCECollector{
		prevFileTimeStamp: "",
		prevLogTimeStamp:  0,
		// TODO : check mcelog process argument, trigger script whether it is avairable metric
		availableMetrics: metrics,
		LogPath:          logPath,
		initialized:      false,
	}
}

type MceLogInfo struct {
	data   map[string]string
	TIME   uint32
	AsItIs string
}

func NewMceLogInfo() *MceLogInfo {
	return &MceLogInfo{
		data:   map[string]string{},
		TIME:   0,
		AsItIs: "",
	}
}

func IsSpecialSymbol(target string) bool {
	// these symbols are not following "KEY VALUE" format
	list := []string{
		"TIME",
		"MCG",
		"MCi",
		"Corrected",
		"Uncorrected",
		"Error",
		"MCi_ADDR",
		"MCA:",
		"CPUID",
	}
	for _, val := range list {
		if target == val {
			return true
		}
	}
	return false
}

func parseMceLogByTime(path string, lastLogTime uint32) ([]MceLogInfo, error) {
	// TODO : return not string, but []???? for each log
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, path+" Open \n")
		return nil, err
	}
	defer file.Close()

	start := false
	sc := bufio.NewScanner(file)
	mcelogs := []MceLogInfo{}
	onelog := NewMceLogInfo()
	for sc.Scan() {
		data := strings.TrimSpace(sc.Text())
		// 1. separate each entry by "Hardware event. This is not a software error."
		if data == "Hardware event. This is not a software error." {
			if start {
				mcelogs = append(mcelogs, *onelog)
			}
			onelog = NewMceLogInfo() // reset
			onelog.AsItIs += data + "\n"
			start = true
			continue
		}
		if !start {
			// reduce nest depth
			continue
		}
		onelog.AsItIs += data + "\n"
		dat := strings.Split(data, " ")
		for i := 0; i < len(dat); i += 2 {
			if IsSpecialSymbol(dat[i]) {
				switch dat[i] {
				case "TIME":
					d, err := strconv.Atoi(dat[i+1])
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
					}
					// 2. compare "TIME 1510397388 Sat Nov 11 19:49:48 2017" lines and previously saved.
					logTime := uint32(d)
					if logTime <= lastLogTime {
						start = false
						break
					}
					onelog.TIME = logTime
					onelog.data["TIMESTR"] = strings.Join(dat[i+1:i+6], " ")
				case "Corrected":
					onelog.data["Corrected"] = "true"
				case "Uncorrected":
					onelog.data["Corrected"] = "false"
				case "CPUID":
					// assumeing this is last line
					onelog.data["CPUID"] = strings.Join(dat[i+1:], " ")
					mcelogs = append(mcelogs, *onelog)
					start = false
				default:
					fmt.Fprintf(os.Stdout, "%s not supported\n", dat[i])
				}
				break
			} else {
				onelog.data[dat[i]] = dat[i+1]
			}
		}
	}
	return mcelogs, nil
}

func (p *MCECollector) GetMceLog() ([]MceLogInfo, error) {
	logs, err := parseMceLogByTime(p.LogPath, p.prevLogTimeStamp)
	if err != nil {
		return nil, err
	}
	if len(logs) > 0 {
		p.prevLogTimeStamp = logs[len(logs)-1].TIME
	}
	return logs, err
}

func (p *MCECollector) WasFileUpdated() (bool, error) {
	fi, err := os.Stat(p.LogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s was not found, did you instll mcelog?\n", p.LogPath)
		return false, err
	}

	modTime := fi.ModTime().String()
	if modTime != p.prevFileTimeStamp {
		p.prevFileTimeStamp = modTime
		return true, nil
	}
	return false, nil
}
