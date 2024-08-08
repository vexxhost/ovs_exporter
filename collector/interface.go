// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"context"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/ovn-org/libovsdb/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/ovs_exporter/ovsmodel"
)

type InterfaceCollector struct {
	logger log.Logger
	ovs    client.Client

	ifaceAdminState                *prometheus.Desc
	ifaceBfdState                  *prometheus.Desc
	ifaceBfdForwarding             *prometheus.Desc
	ifaceBfdRemoteState            *prometheus.Desc
	ifaceBfdFlapCount              *prometheus.Desc
	ifaceStatistics                *prometheus.Desc
	ifaceStatusTunnelEgressCarrier *prometheus.Desc
}

func NewInterfaceCollector(logger log.Logger, ovs client.Client) prometheus.Collector {
	return &InterfaceCollector{
		logger: logger,
		ovs:    ovs,

		ifaceAdminState: prometheus.NewDesc(
			prometheus.BuildFQName("ovs", "interface", "admin_state"),
			"Interface admin state",
			[]string{"name", "mac"},
			nil,
		),
		ifaceBfdState: prometheus.NewDesc(
			prometheus.BuildFQName("ovs", "interface", "bfd_state"),
			"Interface BFD state (-3 = unknown, -2 = admin_down, -1 = down, 0 = init, 1 = up)",
			[]string{"name", "remote_ip"},
			nil,
		),
		ifaceBfdForwarding: prometheus.NewDesc(
			prometheus.BuildFQName("ovs", "interface", "bfd_forwarding"),
			"Interface BFD forwarding",
			[]string{"name", "remote_ip"},
			nil,
		),
		ifaceBfdRemoteState: prometheus.NewDesc(
			prometheus.BuildFQName("ovs", "interface", "bfd_remote_state"),
			"Interface BFD remote state (-3 = unkown, -2 = admin_down, -1 = down, 0 = init, 1 = up)",
			[]string{"name", "remote_ip"},
			nil,
		),
		ifaceBfdFlapCount: prometheus.NewDesc(
			prometheus.BuildFQName("ovs", "interface", "bfd_flap_count"),
			"Interface BFD flap count",
			[]string{"name", "remote_ip"},
			nil,
		),
		ifaceStatistics: prometheus.NewDesc(
			prometheus.BuildFQName("ovs", "interface", "statistics"),
			"Interface statistics",
			[]string{"name", "mac", "type"},
			nil,
		),
		ifaceStatusTunnelEgressCarrier: prometheus.NewDesc(
			prometheus.BuildFQName("ovs", "interface", "status_tunnel_egress_carrier"),
			"Carrier status of the tunnel egress interface",
			[]string{"name", "remote_ip"},
			nil,
		),
	}
}

func (c *InterfaceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ifaceAdminState
	ch <- c.ifaceBfdState
	ch <- c.ifaceBfdForwarding
	ch <- c.ifaceBfdRemoteState
	ch <- c.ifaceBfdFlapCount
	ch <- c.ifaceStatistics
	ch <- c.ifaceStatusTunnelEgressCarrier
}

func (c *InterfaceCollector) Collect(ch chan<- prometheus.Metric) {
	interfaces := &[]ovsmodel.Interface{}
	err := c.ovs.List(context.TODO(), interfaces)
	if err != nil {
		level.Error(c.logger).Log("msg", "Error listing interfaces", "err", err)
		return
	}

	for _, iface := range *interfaces {
		c.collectAdminState(ch, iface)
		c.collectBfdState(ch, iface)
		c.collectBfdForwarding(ch, iface)
		c.collectBfdRemoteState(ch, iface)
		c.collectBfdFlapCount(ch, iface)
		c.collectStatistics(ch, iface)
		c.collectStatusTunnelEgressCarrier(ch, iface)
	}
}

func (c *InterfaceCollector) collectAdminState(ch chan<- prometheus.Metric, iface ovsmodel.Interface) {
	adminState := float64(0)
	if iface.AdminState != nil && *iface.AdminState == "up" {
		adminState = 1
	}

	mac := "unknown"
	if iface.MACInUse != nil {
		mac = *iface.MACInUse
	} else {
		level.Warn(c.logger).Log("msg", "MAC address is nil for interface", "interface", iface.Name)
	}

	ch <- prometheus.MustNewConstMetric(
		c.ifaceAdminState,
		prometheus.GaugeValue,
		adminState,
		iface.Name,
		mac,
	)
}

func (c *InterfaceCollector) collectBfdState(ch chan<- prometheus.Metric, iface ovsmodel.Interface) {
	_, ok := iface.BFDStatus["state"]
	if !ok {
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.ifaceBfdState,
		prometheus.GaugeValue,
		mapBfdState(iface.BFDStatus["state"]),
		iface.Name,
		iface.Options["remote_ip"],
	)
}

func (c *InterfaceCollector) collectBfdForwarding(ch chan<- prometheus.Metric, iface ovsmodel.Interface) {
	_, ok := iface.BFDStatus["forwarding"]
	if !ok {
		return
	}

	bfdForwarding := float64(0)
	if iface.BFDStatus["forwarding"] == "true" {
		bfdForwarding = 1
	}

	ch <- prometheus.MustNewConstMetric(
		c.ifaceBfdForwarding,
		prometheus.GaugeValue,
		bfdForwarding,
		iface.Name,
		iface.Options["remote_ip"],
	)
}

func (c *InterfaceCollector) collectBfdRemoteState(ch chan<- prometheus.Metric, iface ovsmodel.Interface) {
	_, ok := iface.BFDStatus["remote_state"]
	if !ok {
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.ifaceBfdRemoteState,
		prometheus.GaugeValue,
		mapBfdState(iface.BFDStatus["remote_state"]),
		iface.Name,
		iface.Options["remote_ip"],
	)
}

func (c *InterfaceCollector) collectBfdFlapCount(ch chan<- prometheus.Metric, iface ovsmodel.Interface) {
	_, ok := iface.BFDStatus["flap_count"]
	if !ok {
		return
	}

	flapCount, err := strconv.ParseFloat(iface.BFDStatus["flap_count"], 64)
	if err != nil {
		level.Error(c.logger).Log("msg", "Error parsing BFD flap count", "err", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.ifaceBfdFlapCount,
		prometheus.GaugeValue,
		flapCount,
		iface.Name,
		iface.Options["remote_ip"],
	)
}

func (c *InterfaceCollector) collectStatistics(ch chan<- prometheus.Metric, iface ovsmodel.Interface) {
	mac := "unknown"
	if iface.MACInUse != nil {
		mac = *iface.MACInUse
	} else {
		level.Warn(c.logger).Log("msg", "MAC address is nil for interface", "interface", iface.Name)
	}

	for stat, value := range iface.Statistics {
		ch <- prometheus.MustNewConstMetric(
			c.ifaceStatistics,
			prometheus.CounterValue,
			float64(value),
			iface.Name,
			mac,
			stat,
		)
	}
}

func (c *InterfaceCollector) collectStatusTunnelEgressCarrier(ch chan<- prometheus.Metric, iface ovsmodel.Interface) {
	_, ok := iface.Status["tunnel_egress_iface_carrier"]
	if !ok {
		return
	}

	statusTunnelEgressCarrier := float64(0)
	if iface.Status["tunnel_egress_iface_carrier"] == "up" {
		statusTunnelEgressCarrier = 1
	}

	ch <- prometheus.MustNewConstMetric(
		c.ifaceStatusTunnelEgressCarrier,
		prometheus.GaugeValue,
		statusTunnelEgressCarrier,
		iface.Name,
		iface.Options["remote_ip"],
	)
}

func mapBfdState(state string) float64 {
	switch state {
	case "admin_down":
		return -2
	case "down":
		return -1
	case "init":
		return 0
	case "up":
		return 1
	default:
		return -3
	}
}
