// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package tcpcheckreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcpcheckreceiver"

import (
	"context"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcpcheckreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	collectorscraper "go.opentelemetry.io/collector/scraper"
)

// NewFactory creates a factory for tcpcheckreceiver receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		newDefaultConfig,
		receiver.WithMetrics(newReceiver, metadata.MetricsStability))
}

func newDefaultConfig() component.Config {
	cfg := scraperhelper.NewDefaultControllerConfig()

	return &Config{
		ControllerConfig:     cfg,
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
		Targets:              []*confignet.TCPAddrConfig{},
	}
}

func newReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	tcpCheckConfig, ok := cfg.(*Config)
	if !ok {
		return nil, errConfigTCPCheck
	}

	mp := newScraper(tcpCheckConfig, settings)
	s, err := collectorscraper.NewMetrics(mp.scrape)
	if err != nil {
		return nil, err
	}
	opt := scraperhelper.AddScraper(metadata.Type, s)

	return scraperhelper.NewScraperControllerReceiver(
		&tcpCheckConfig.ControllerConfig,
		settings,
		consumer,
		opt,
	)
}

// timeout
//func createDefaultConfig() component.Config {
//	cfg := scraperhelper.NewDefaultControllerConfig()
//	cfg.CollectionInterval = 10 * time.Second
//
//	return &Config{
//		ControllerConfig: cfg,
//		TCPClientSettings: configtcp.TCPClientSettings{
//			Timeout: 10 * time.Second,
//		},
//		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
//	}
//}

//func createMetricsReceiver(_ context.Context, params receiver.Settings, rConf component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
//	cfg, ok := rConf.(*Config)
//	if !ok {
//		return nil, errConfigNotTCPCheck
//	}
//
//	tcpCheckScraper := newScraper(cfg, params)
//	scraper, err := scraperhelper.NewScraper(metadata.Type, tcpCheckScraper.scrape, scraperhelper.WithStart(tcpCheckScraper.start))
//	if err != nil {
//		return nil, err
//	}
//
//	return scraperhelper.NewScraperControllerReceiver(&cfg.ControllerConfig, params, consumer, scraperhelper.AddScraper(scraper))
//}
