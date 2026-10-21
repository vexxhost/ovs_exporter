# Copyright (c) 2024 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.27.3@sha256:f20d83a83e18dbd2a4714b8ffd4b30b287b6243962dd75608e86abcc63ba4efc AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /ovs_exporter

FROM scratch
COPY --from=builder /ovs_exporter /bin/ovs_exporter
EXPOSE 9180
ENTRYPOINT ["/bin/ovs_exporter"]
