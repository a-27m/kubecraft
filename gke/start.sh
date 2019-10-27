#!/bin/sh

gcloud container clusters create kubecraft
kubectl apply -f ../manifests/
kubectl get svc kubecraft
