#!/bin/sh
docker build --rm --no-cache --tag eu.gcr.io/mauve-cloud/cni-health-probe .
docker push eu.gcr.io/mauve-cloud/cni-health-probe

