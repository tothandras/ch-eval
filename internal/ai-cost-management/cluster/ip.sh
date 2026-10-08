#!/usr/bin/env bash
# ip.sh <service> -> container IP on the compose network
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "aicm-3x2-$1-1"
