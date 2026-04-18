# Open vSwitch Exporter

## Metrics

### Interface Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ovs_interface_admin_state` | Gauge | `name`, `mac` | Interface admin state (0 = down, 1 = up) |
| `ovs_interface_bfd_state` | Gauge | `name`, `remote_ip` | Interface BFD state (-3 = unknown, -2 = admin_down, -1 = down, 0 = init, 1 = up) |
| `ovs_interface_bfd_forwarding` | Gauge | `name`, `remote_ip` | Interface BFD forwarding (0 = false, 1 = true) |
| `ovs_interface_bfd_remote_state` | Gauge | `name`, `remote_ip` | Interface BFD remote state (-3 = unknown, -2 = admin_down, -1 = down, 0 = init, 1 = up) |
| `ovs_interface_bfd_flap_count` | Gauge | `name`, `remote_ip` | Interface BFD flap count |
| `ovs_interface_statistics` | Counter | `name`, `mac`, `type` | Interface statistics (rx_bytes, tx_bytes, rx_packets, tx_packets, rx_errors, tx_errors, rx_dropped, tx_dropped, etc.) |
| `ovs_interface_status_tunnel_egress_carrier` | Gauge | `name`, `remote_ip` | Carrier status of the tunnel egress interface (0 = down, 1 = up) |
