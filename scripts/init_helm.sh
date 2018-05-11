#!/bin/sh

if [ -f /app/config/catalog.yaml ]; then
    cp /app/config/catalog.yaml /app/
fi

/app/helmi -config /app/catalog.yaml