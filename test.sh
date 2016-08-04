#!/bin/bash

while :
do
  IP=$(curl localhost:9090/metrics)
  curl -i -XPOST   'http://localhost:8086/write?db=mydb' --data-binary "$IP"
  sleep 1
done
