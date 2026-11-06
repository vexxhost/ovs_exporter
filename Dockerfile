# Copyright (c) 2024 VEXXHOST, Inc.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.27.3@sha256:e519d3b0a5b285e5aa53ab9eef02bc6bbb9cb04608389b743a136a9ce75f0d1e AS builder
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /ovs_exporter

FROM scratch
COPY --from=builder /ovs_exporter /bin/ovs_exporter
EXPOSE 9180
ENTRYPOINT ["/bin/ovs_exporter"]
