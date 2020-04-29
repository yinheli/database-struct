#!/usr/bin/env bash
set -e

cd $(dirname $0)

pkger -o /pkg/static -include /assets -include /template
