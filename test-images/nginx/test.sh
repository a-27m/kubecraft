#!/bin/sh

PODS=`kubectl get pods --selector='run=nginx' -o name`

for pod in $PODS; do
  RESP=`kubectl exec -t $pod -- curl -Is http://127.0.0.1 | head -n1`
  echo "$pod: $RESP"
done
