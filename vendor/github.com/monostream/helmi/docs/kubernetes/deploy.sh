#!/bin/bash

# deploy helmi
kubectl create -f $(dirname $(readlink --canonicalize-existing "$0"))/kube-helmi-rbac.yaml
kubectl create -f $(dirname $(readlink --canonicalize-existing "$0"))/kube-helmi-secret.yaml
kubectl create -f $(dirname $(readlink --canonicalize-existing "$0"))/kube-helmi.yaml