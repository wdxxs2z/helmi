#!/bin/sh

curl -ss "http://localhost:5000/v2/catalog" | json_pp
