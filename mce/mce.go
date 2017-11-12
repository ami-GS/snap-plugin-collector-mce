package mce

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
		mceLogs, err := getMceLog(mceLog, p.prevLogTimeStamp)
		if err != nil {
			return nil, err
		}
		p.prevLogTimeStamp = mceLogs[len(mceLogs)-1].TIME
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

type mceLogFormat struct {
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
	ADDR      uint64
	TIME      uint32
	TIMESTR   string
	MCG       string
	MCi       string
	Corrected bool
	Error     string //???enabled?
	MCiADDR   string //???
	STATUS    uint64
	MCGSTATUS uint16
	MCGCAP    uint32
	APICID    uint16
	SOCKETID  uint8
	CPUID     string
	AsItIs    string // will be removed, for /everything
}

func getMceLog(path string, lastLogTime uint32) ([]mceLogFormat, error) {
	// TODO : return not string, but []???? for each log
	file, err := os.Open(mceLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, mceLog+" Open \n")
		return nil, err
	}
	defer file.Close()

	start := false
	sc := bufio.NewScanner(file)
	mcelogs := []mceLogFormat{}
	onelog := mceLogFormat{}
	for sc.Scan() {
		data := sc.Text()
		// 1. separate each entry by "Hardware event. This is not a software error."
		if data == "Hardware event. This is not a software error." {
			if start {
				mcelogs = append(mcelogs, onelog)
			}
			onelog = mceLogFormat{} // reset
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
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.MCE = uint8(d)
			case "CPU":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.CPU = uint8(d)
			case "BANK":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.BANK = uint8(d)
			case "ADDR":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.ADDR = uint64(d)
			case "TIME":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				// 2. compare "TIME 1510397388 Sat Nov 11 19:49:48 2017" lines and previously saved.
				logTime := uint32(d)
				if logTime <= lastLogTime {
					start = false
					break
				}
				onelog.TIME = logTime
				onelog.TIMESTR = strings.Join(dat[i+1:i+6], " ")
			case "STATUS":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.STATUS = uint64(d)
			case "MCGSTATUS":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.MCGSTATUS = uint16(d)
			case "MCGCAP":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.MCGCAP = uint32(d)
			case "APICID":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.APICID = uint16(d)
			case "SOCKETID":
				d, err := strconv.Atoi(dat[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s format error", dat[i])
				}
				onelog.SOCKETID = uint8(d)
			case "CPUID":
				// assumeing this is last line
				onelog.CPUID = strings.Join(dat[i+1:], " ")
				start = false
			case "MCG":
			case "MCi":
			case "MCA":
			default:
				fmt.Fprintf(os.Stdout, "not supported")
			}
		}
	}
	return mcelogs, nil
}
