#!/bin/bash

SCRIPT=$(readlink -f "${0}")
DIR=$(dirname "${SCRIPT}")
TAG="wimp:latest"

docker build -t "${TAG}" "${DIR}"

echo "run with: 
    ${DIR}/server.sh
    ${DIR}/server.sh -e WIMP=true -e ENDPOINT=/endpoint"
