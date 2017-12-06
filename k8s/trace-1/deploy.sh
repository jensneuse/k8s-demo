#!/usr/bin/env bash

namespace="trace-1"

for file in Data/*.yaml
do
    name=${file##*/}
    istioctl kube-inject -f $name | kubectl -n $namespace apply -f -
done