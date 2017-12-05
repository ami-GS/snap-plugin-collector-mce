package mceStream

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ami-GS/snap-plugin-collector-mce/mce"
	"github.com/go-fsnotify/fsnotify"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type MCEStreamCollector struct {
	base       *mce.MCECollector
	rcvMetrics []plugin.Metric
}

func NewStream(logPath string) *MCEStreamCollector {
	// TODO : logPath should be configurable, like mce.go
	return &MCEStreamCollector{
		mce.New(logPath),
		nil,
	}
}

func (p *MCEStreamCollector) StreamMetrics(ctx context.Context, metricsIn chan []plugin.Metric, metricsOut chan []plugin.Metric, err chan string) error {
	go p.msgSender(metricsOut, err)
	p.msgReceiver(metricsIn)
	return nil
}

func (p *MCEStreamCollector) msgReceiver(metricsIn chan []plugin.Metric) {
	for {
		var mts []plugin.Metric
		mts = <-metricsIn
		p.rcvMetrics = mts
	}
}

func (p *MCEStreamCollector) msgSender(metricsOut chan []plugin.Metric, errOut chan string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		errOut <- err.Error()
	}
	defer watcher.Close()

	// parse log, then send
	parseAndSend := func() {
		mceLogs, err := p.base.GetMceLog()
		if err != nil {
			errOut <- fmt.Sprintf("issue when opening %s", p.base.LogPath)
		}
		metricsOut <- mce.StuffLogToMetrics(mceLogs, p.rcvMetrics)
	}

	// event based sender
	go func() {
		for {
			select {
			case <-watcher.Events:
				parseAndSend()
			case err = <-watcher.Errors:
				errOut <- err.Error()
			}
		}
	}()
	if err = watcher.Add(p.base.LogPath); err != nil {
		errOut <- err.Error()
	}

	// request based sender
	for {
		if p.rcvMetrics == nil {
			time.Sleep(time.Second)
			continue
		}
		recvMetrics := p.rcvMetrics

		if recvMetrics == nil {
			// TODO : this time should be configurable
			time.Sleep(time.Second)
			continue
		}

		ok, err := p.base.WasFileUpdated()
		if err != nil {
			errOut <- fmt.Sprintf("issue when opening %s", p.base.LogPath)
		}
		if ok {
			parseAndSend()
		}
		// TODO : this should be configurable
		time.Sleep(time.Millisecond * time.Duration(rand.Int63n(1000)))
	}
}

func (p *MCEStreamCollector) GetMetricTypes(tmp plugin.Config) ([]plugin.Metric, error) {
	return p.base.GetMetricTypes(tmp)
}

func (p *MCEStreamCollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	return p.base.GetConfigPolicy()
}
