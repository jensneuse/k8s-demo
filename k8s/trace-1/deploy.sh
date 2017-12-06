#!/usr/bin/env bash

namespace="trace-1"

kubectl -n $namespace apply -f <(istioctl kube-inject -f database-deployment.yaml)
database-service.yaml
kubectl -n $namespace apply -f <(istioctl kube-inject -f horrorskope-api-deployment.yaml)
horrorskope-api-service.yaml
kubectl -n $namespace apply -f <(istioctl kube-inject -f horrorskope-external-api-deployment.yaml)
horrorskope-external-api-service.yaml
kubectl -n $namespace apply -f <(istioctl kube-inject -f renderine-deployment.yaml)
renderine-ingress.yaml
renderine-service.yaml
kubectl -n $namespace apply -f <(istioctl kube-inject -f weather-api-deployment.yaml)
weather-api-service.yaml
kubectl -n $namespace apply -f <(istioctl kube-inject -f weather-external-api-deployment.yaml)
weather-external-api-service.yaml