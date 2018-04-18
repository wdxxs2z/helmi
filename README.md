# helmi
Open Service Broker API Implementation using helm &amp; kubectl

![alt Logo](docs/logo.png)

## Start locally

```console
# start minikube
minikube start

# init helm and install tiller (needed once)
helm init

# build helmi
go get -d github.com/monostream/helmi
cd ${GOPATH:-~/go}/src/github.com/monostream/helmi
go build

# run helmi
./helmi
```

## Start on kubernetes

```console
# create serviceaccount, clusterrolebinding, deployment, service and an optional secret for basic authorization
kubectl create -f docs/kubernetes/kube-helmi-rbac.yaml
kubectl create -f docs/kube-helmi-secret.yaml
kubectl create -f docs/kubernetes/kube-helmi.yaml

# curl to catalog with basic auth
curl --user {username}:{password} http://$(kubernetes ip):30000/v2/catalog
```
or
```console
./docs/kubernetes/deploy.sh

# curl to catalog with basic auth
curl --user {username}:{password} http://$(kubernetes ip):30000/v2/catalog
```

## Start with kube helm

Configure the values.

| Parameter               | Description                            | Default                   |
| ----------------------- | -------------------------------------- | ------------------------- |
| `helmi.username`|This is the helmi service broker username.|admin|
| `helmi.password`|This is the helmi service broker password.|helmi|
| `helmi.repo_url`|This is the chart repo url.|""|
| `helmi.repo_name`|This is the chart repo name.|""|
| `ingress.hosts`|The helmi ingress hosts.|""|
| `kubeconfig.*`|Must set the kubeconfig.|""|
| `tls.cacert`|Must set the kube ca cert.|""|

Install the helmi release.

```
helm install -n helmi-core --namespace helmi-system .
```

Test the helmi url with ingress.

```
curl -k 'https://helmi-service-broker.k8s.io/v2/catalog' -i -X GET \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     -u 'admin:helmi'
```

Create mariadb instance.
```
curl -k 'https://helmi-service-broker.k8s.io/v2/service_instances/3b2e7d2c915242a5befcf03e1c3f47cd' -X PUT \
-H 'Accept: application/json' \
-H 'Content-Type: application/x-www-form-urlencoded' \
-u 'admin:helmi' \
-d '{"service_instance_guid":"a0029c76-7017-4a74-94b0-54a04ad94b80","plan_id":"e79306ef-4e10-4e3d-b38e-ffce88c90f59","service_id":"ab53df4d-c279-4880-94f7-65e7d72b7834","app_guid":"081d55a0-1bfa-4e51-8d08-273f764988db","context": {"platform":"kubernetes","namespace":"mariadb-test"},"parameters":{"serviceType":"NodePort"},"name":"mariadb-service"}'
```

Delete mariadb install.

```
curl -k 'https://helmi-service-broker.k8s.io/v2/service_instances/3b2e7d2c915242a5befcf03e1c3f47cd' -i -X DELETE \
-H 'Accept: application/json' \
-H 'Content-Type: application/x-www-form-urlencoded' \
-u 'admin:helmi'
```
## Use in Kubernetes

If we install it with helm, we can list the service broker on kubernetes

```
# kubectl get clusterservicebroker
NAME                   URL
helmi-service-broker   http://helmi-service-broker.k8s.io
```

List the service class

```
# kubectl get clusterserviceclasses
NAME                                   EXTERNAL NAME      BROKER                 BINDABLE   PLAN UPDATABLE
201cb950-e640-4453-9d91-4708ea0a1342   cassandra          helmi-service-broker   true      false
2f1e7c63-0511-4209-aa7f-6bdee7ffb2b6   rabbitmq           helmi-service-broker   true      false
777f5478-5796-426a-ab8a-5d3dc5e1bdcc   muescheli          helmi-service-broker   true      false
8dda5a6f-f796-4b52-806f-4129d7576d6e   minio              helmi-service-broker   true      false
ab53df4d-c279-4880-94f7-65e7d72b7834   mariadb            helmi-service-broker   true      false
b4280104-b578-4156-a69c-8961bcdfa8c0   mongodb            helmi-service-broker   true      false
c26e6c7a-fe17-4568-ac4c-46545ab1d178   redis              helmi-service-broker   true      false
```

List the service plans

```
# kubectl get clusterserviceplans
NAME                                   EXTERNAL NAME    BROKER                 CLASS
169d5466-12c9-4a89-a063-f72048b3d4c4   free             helmi-service-broker   201cb950-e640-4453-9d91-4708ea0a1342
381c8dd1-676b-4d1f-ae00-97e8304f966f   free             helmi-service-broker   c26e6c7a-fe17-4568-ac4c-46545ab1d178
75b7b1de-70ef-4499-b55c-e2337d320626   free             helmi-service-broker   777f5478-5796-426a-ab8a-5d3dc5e1bdcc
7b16d6aa-260a-4b8d-b12c-464d2cedb9d0   dev              helmi-service-broker   201cb950-e640-4453-9d91-4708ea0a1342
905b1f0e-c815-41d4-b3e4-6ccb602b9e8e   free             helmi-service-broker   b4280104-b578-4156-a69c-8961bcdfa8c0
d2badac0-8e41-4588-a9fc-0e662c480610   free             helmi-service-broker   2f1e7c63-0511-4209-aa7f-6bdee7ffb2b6
e79306ef-4e10-4e3d-b38e-ffce88c90f59   free             helmi-service-broker   ab53df4d-c279-4880-94f7-65e7d72b7834
f003f191-c250-4e85-9abd-038af629ad71   free             helmi-service-broker   8dda5a6f-f796-4b52-806f-4129d7576d6e
```

Create service instance with parameters

```
```

Bind a service instance

```
```

## Use in Cloud Foundry

Register Helmi Service Broker

```console
cf create-service-broker helmi {username} {password} http://{IP}:5000
```

List and allow service access

```console
cf service-access
cf enable-service-access {service}
```

List marketplace and create service instance

```console
cf marketplace
cf create-service {service} {plan} {name}
```

Bind service to application

```console
cf bind-service {app} {name}
```

## Tests
run tests
```console
go test ./pkg/* -v
```

## Environment Variables

Helmi can use environment variables to define a dns name for connection strings and a username/password for basic authentication.

To use basic authentication set `USERNAME` and `PASSWORD` environment variables. In the k8s deployment they are read from a secret, see [kube-helmi-secret.yaml](docs/kubernetes/kube-helmi-secret.yaml)

To replace the connection string IPs set an environment variable `DOMAIN`.