#!/bin/bash

for i in `seq 10`; do 
    (curl -s localhost:8080/query >/dev/null &)
     sleep 0.2
done
