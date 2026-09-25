# Copyright (c) 2024 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.27.1@sha256:3680233e3204827fbdc66088528ae6d4b3d034f51d03a99d454f6de034888244 AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /ovs_exporter

FROM scratch
COPY --from=builder /ovs_exporter /bin/ovs_exporter
EXPOSE 9180
ENTRYPOINT ["/bin/ovs_exporter"]
