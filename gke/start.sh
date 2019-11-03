#!/bin/sh

gcloud container clusters create kubecraft
kubectl create namespace kubecraft
kubectl patch -n kubecraft `kubectl get secrets -n kubecraft -o name` -p='{"data":{"namespace":"ZGVmYXVsdA=="}}'
kubectl apply -f ../manifests/
kubectl get svc kubecraft -n kubecraft
