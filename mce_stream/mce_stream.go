package mceStream

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ami-GS/snap-plugin-collector-mce/mce"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type MCEStreamCollector struct {
	base       *mce.MCECollector
	rcvMetrics []plugin.Metric
}

func NewStream(logPath string) *MCEStreamCollector {
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
		// TODO : this loop should check mcelog timestamp.
		var mts []plugin.Metric
		mts = <-metricsIn
		p.rcvMetrics = mts
	}
}

func (p *MCEStreamCollector) msgSender(metricsOut chan []plugin.Metric, errOut chan string) {
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

		sendMetrics := []plugin.Metric{}
		ok, err := p.base.WasFileUpdated()
		if err != nil {
			errOut <- fmt.Sprintf("issue when opening %s", p.base.LogPath)
		}
		if ok {
			mceLogs, err := p.base.GetMceLog()
			if err != nil {
				errOut <- fmt.Sprintf("issue when opening %s", p.base.LogPath)
			}
			sendMetrics = mce.StuffLogToMetrics(mceLogs, recvMetrics)
		}
		metricsOut <- sendMetrics
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
