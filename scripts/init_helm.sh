#!/bin/sh

if [ -f /app/config/catalog.yaml ]; then
    sed -i 's/\\{\\{/{{/g' /app/config/catalog.yaml
    sed -i 's/\\}\\}/}}/g' /app/config/catalog.yaml
    cp /app/config/catalog.yaml /app/
fi

/app/helmi -config /app/catalog.yaml