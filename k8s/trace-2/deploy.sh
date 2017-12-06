#!/usr/bin/env bash

namespace="trace-2"

istioctl kube-inject -f database-deployment.yaml | kubectl -n $namespace apply -f -
kubectl -n $namespace apply -f database-service.yaml
istioctl kube-inject -f horrorskope-api-deployment.yaml | kubectl -n $namespace apply -f -
kubectl -n $namespace apply -f horrorskope-api-service.yaml
istioctl kube-inject -f horrorskope-external-api-deployment.yaml | kubectl -n $namespace apply -f -
kubectl -n $namespace apply -f horrorskope-external-api-service.yaml
istioctl kube-inject -f renderine-deployment.yaml | kubectl -n $namespace apply -f -
kubectl -n $namespace apply -f renderine-ingress.yaml
kubectl -n $namespace apply -f renderine-service.yaml
istioctl kube-inject -f weather-api-deployment.yaml | kubectl -n $namespace apply -f -
kubectl -n $namespace apply -f weather-api-service.yaml
istioctl kube-inject -f weather-external-api-deployment.yaml | kubectl -n $namespace apply -f -
kubectl -n $namespace apply -f weather-external-api-service.yaml