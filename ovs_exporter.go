// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/ovn-org/libovsdb/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/tonglil/gokitlogr"

	"github.com/vexxhost/ovs_exporter/collector"
	"github.com/vexxhost/ovs_exporter/ovsmodel"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	ovsdbEndpoint = kingpin.Flag(
		"ovsdb.endpoint",
		"Endpoint for OVSDB",
	).Envar("OVSDB_ENDPOINT").Default("unix:/var/run/openvswitch/db.sock").String()
	toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9272")
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)

	kingpin.Version(version.Print("ovs_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)
	logr := gokitlogr.New(&logger)

	level.Info(logger).Log("msg", "Starting ovs_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	dbModelReq, err := ovsmodel.FullDatabaseModel()
	if err != nil {
		level.Error(logger).Log("msg", "Error getting OVSDB model", "err", err)
		os.Exit(1)
	}

	ovs, err := client.NewOVSDBClient(dbModelReq, client.WithEndpoint(*ovsdbEndpoint), client.WithLogger(&logr))
	if err != nil {
		level.Error(logger).Log("msg", "Error creating OVSDB client", "err", err)
		os.Exit(1)
	}

	err = ovs.Connect(context.Background())
	if err != nil {
		level.Error(logger).Log("msg", "Error connecting to OVSDB", "err", err)
		os.Exit(1)
	}
	defer ovs.Close()

	ovs.MonitorAll(context.TODO())

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collector.NewInterfaceCollector(logger, ovs),
	)

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Open vSwitch Exporter",
			Description: "Prometheus Exporter for Open vSwitch",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
