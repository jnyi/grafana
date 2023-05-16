#!/bin/bash

make build-docker-full && \
docker tag grafana/grafana:dev registry.dev.databricks.com/grafana-9.1.x-base-alpine-3.17:latest && \
docker push registry.dev.databricks.com/grafana-9.1.x-base-alpine-3.17:latest

