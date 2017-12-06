#!/usr/bin/env bash

namespace="trace-1"

for file in Data/*.yaml
do
    name=${file##*/}
    kubectl -n $namespace apply -f <(istioctl kube-inject -f $name)
done