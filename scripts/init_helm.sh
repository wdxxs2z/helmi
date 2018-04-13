#!/bin/sh

if [ "$HELM_REPO_URL" != "" ] && [ "$HELM_REPO_NAME" != "" ]; then
    helm init --client-only && \
    helm repo add $HELM_REPO_NAME $HELM_REPO_URL && \
    helm repo update
else
    helm init --client-only && \
    helm repo update
fi

if [ -f /app/config/catalog.yaml ]; then
    cp /app/config/catalog.yaml /app/
fi

ls -laht
/app/helmi