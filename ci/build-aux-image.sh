#!/bin/bash

TOKEN=$(curl -s "https://auth.docker.io/token?service=registry.docker.io&scope=repository:$DOCKER_AUX_REPO:pull" | jq -r .token)
DATE=$(curl -H "Authorization: Bearer $TOKEN" https://index.docker.io/v2/$DOCKER_AUX_REPO/manifests/latest 2>/dev/null \
    | jq -r '.history[].v1Compatibility' \
    | jq '.created' \
    | sort \
    | tail -n1)
CURRENT=$(echo $DATE | grep $(date -u +%Y-%m-%dT))

if [ -z "$CURRENT" ]; then
    docker build -t $DOCKER_AUX_REPO -f ci/aux/Dockerfile .
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    docker push $DOCKER_AUX_REPO
else
    echo "Aux image is already current"
fi
