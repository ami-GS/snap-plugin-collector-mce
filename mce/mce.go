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

var MceLogPath = "/var/log/mcelog"

const metricAll string = "everything"

// AllMetricsNames : for first testing
var AllMetricsNames = []string{
	// TODO : enable CPU/0, CPU/1, ADDR/0x1234, like metrics
	"CPU",
	"ADDR",
	"BANK",
	"Corrected",
	"Uncorrected",
}

type MCECollector struct {
	// this is used for checking file change
	prevFileTimeStamp string
	// this is used for picking up latest log
	prevLogTimeStamp uint32
	// this is decided by mcelog process argument
	availableMetrics []string
	// this would be mceLog
	logPath string
}

func (p *MCECollector) GetMetricTypes(_ plugin.Config) ([]plugin.Metric, error) {
	metricTypes := []plugin.Metric{}
	for i := 0; i < len(p.availableMetrics); i++ {
		metricType := plugin.Metric{
			Namespace: plugin.NewNamespace(VendorName, PluginName, p.availableMetrics[i]),
		}
		metricTypes = append(metricTypes, metricType)
	}
	return metricTypes, nil
}

func (p *MCECollector) CollectMetrics(metricTypes []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	ok, err := p.WasFileUpdated()
	if err != nil {
		return nil, err
	}
	if ok {
		ts := time.Now()
		mceLogs, err := p.GetMceLog()
		if err != nil {
			return nil, err
		}
		// TODO : need to consider how to manage several logs
		data := ""
		for _, mceLog := range mceLogs {
			data += mceLog.AsItIs + "\n"
		}

		for _, metricType := range metricTypes {
			ns := metricType.Namespace
			metric := plugin.Metric{
				Namespace: ns,
				Data:      data, // TODO : use appropriate telemetry
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
	policy.AddNewStringRule([]string{VendorName, "???", PluginName}, "key", false, plugin.SetDefaultString(p.logPath))
	return *policy, nil
}

// New creates instance of interface info plugin
func New(logPath string) *MCECollector {
	metrics := []string{"cpu", "memory", metricAll}
	return &MCECollector{
		prevFileTimeStamp: "",
		prevLogTimeStamp:  0,
		// TODO : check mcelog process argument, trigger script whether it is avairable metric
		availableMetrics: metrics,
		logPath:          logPath,
	}
}

type MceLogFormat struct {
	/*
	   MCE 0
	   CPU 1 BANK 2
	   ADDR 1234
	   TIME 1510397270 Sat Nov 11 19:47:50 2017
	   MCG status:
	   MCi status:
	   Corrected error
	   Error enabled
	   MCi_ADDR register valid
	   MCA: No Error
	   STATUS 9400000000000000 MCGSTATUS 0
	   MCGCAP 7000c16 APICID 2 SOCKETID 0
	   CPUID Vendor Intel Family 6 Model 79
	*/
	MCE       uint8
	CPU       uint8
	BANK      uint8
	MISC      uint16
	ADDR      string // temporaly
	TIME      uint32
	TIMESTR   string
	MCG       string
	MCi       string
	Corrected bool
	Error     string //???enabled?
	MCiMISC   string //???
	MCiADDR   string //???
	MCA       string
	CACHE     string
	STATUS    uint64
	MCGSTATUS uint16
	MCGCAP    string // temporaly
	APICID    uint16
	SOCKETID  uint8
	CPUID     string
	AsItIs    string // will be removed, for /everything
}

func (p *MCECollector) GetMceLog() ([]MceLogFormat, error) {
	logs, err := parseMceLogByTime(p.logPath, p.prevLogTimeStamp)
	if err != nil {
		return nil, err
	}
	if len(logs) > 0 {
		p.prevLogTimeStamp = logs[len(logs)-1].TIME
	}
	return logs, err
}

func parseMceLogByTime(path string, lastLogTime uint32) ([]MceLogFormat, error) {
	// TODO : return not string, but []???? for each log
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, path+" Open \n")
		return nil, err
	}
	defer file.Close()

	start := false
	sc := bufio.NewScanner(file)
	mcelogs := []MceLogFormat{}
	onelog := MceLogFormat{}
	for sc.Scan() {
		data := strings.TrimSpace(sc.Text())
		// 1. separate each entry by "Hardware event. This is not a software error."
		if data == "Hardware event. This is not a software error." {
			if start {
				mcelogs = append(mcelogs, onelog)
			}
			onelog = MceLogFormat{} // reset
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
			switch dat[i] {
			case "MCE":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.MCE = uint8(d)
			case "CPU":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.CPU = uint8(d)
			case "BANK":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.BANK = uint8(d)
			case "MISC":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.MISC = uint16(d)
			case "ADDR":
				onelog.ADDR = dat[i+1]
			case "TIME":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				// 2. compare "TIME 1510397388 Sat Nov 11 19:49:48 2017" lines and previously saved.
				logTime := uint32(d)
				if logTime <= lastLogTime {
					start = false
					goto OUT
				}
				onelog.TIME = logTime
				onelog.TIMESTR = strings.Join(dat[i+1:i+6], " ")
				goto OUT
			case "Corrected":
				onelog.Corrected = true
			case "Uncorrected":
				onelog.Corrected = false
			case "STATUS":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.STATUS = uint64(d)
			case "MCGSTATUS":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.MCGSTATUS = uint16(d)
			case "MCGCAP":
				onelog.MCGCAP = dat[i+1]
			case "APICID":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.APICID = uint16(d)
			case "SOCKETID":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error: %s\n", dat[i], dat[i+1])
				}
				onelog.SOCKETID = uint8(d)
			case "CPUID":
				// assumeing this is last line
				onelog.CPUID = strings.Join(dat[i+1:], " ")
				mcelogs = append(mcelogs, onelog)
				start = false
				goto OUT
			case "MCA:":
				onelog.MCA = strings.Join(dat[i+1:], " ")
				goto OUT
			case "Generic":
				onelog.CACHE = strings.Join(dat[i+i:], " ")
				goto OUT
			case "MCi_ADDR":
				onelog.MCiADDR = strings.Join(dat[i+1:], " ")
				goto OUT
			case "MCG":
			case "MCi":
			default:
				fmt.Fprintf(os.Stdout, "%s not supported\n", dat[i])
			}
		}
	OUT:
	}

	return mcelogs, nil
}

func (p *MCECollector) WasFileUpdated() (bool, error) {
	fi, err := os.Stat(p.logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s was not found, did you instll mcelog?\n", p.logPath)
		return false, err
	}

	modTime := fi.ModTime().String()
	if modTime != p.prevFileTimeStamp {
		p.prevFileTimeStamp = modTime
		return true, nil
	}
	return false, nil
}
