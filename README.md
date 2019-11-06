# istio-mixer-adapter

First of all, get go environment ready, and docker, kubernetes, istio as well.

## Prepare Development Environment

~~~ shell
export GOPATH=$HOME/go

# checkout the adapter source code
cd /tmp
git clone https://github.com/jianshaow/istio-mixer-adapter.git
export ADAPTER_REPO=/tmp/istio-mixer-adapter

# checkout istio source code
mkdir -p $GOPATH/src/istio.io/ && \
cd $GOPATH/src/istio.io/  && \
git clone https://github.com/istio/istio
# base on stable version
cd istio
git checkout 1.2.9

# set the environment variable
export ISTIO=$GOPATH/src/istio.io
export MIXER_REPO=$GOPATH/src/istio.io/istio/mixer

# build mixer server and client
pushd $ISTIO/istio && make mixs
pushd $ISTIO/istio && make mixc

~~~

## Authorization Adapter

### Generate Code

~~~ shell

# copy adapter source code into istio mixer repo
cp $ADAPTER_REPO/authzadapter $MIXER_REPO/adapter/ -r

# generate config gRPC code
go generate $MIXER_REPO/adapter/authzadapter/
# go generate needs docker, may need root privilege, if so, GOPATH needs to be passed for sudo as following
# sudo GOPATH=$GOPATH go generate $MIXER_REPO/adapter/authzadapter/
go build $MIXER_REPO/adapter/authzadapter/

~~~

### Test with Mixer Server & Client

~~~ shell

# copy generated adapter manifest
mkdir $ADAPTER_REPO/authzadapter/testdata
cp $MIXER_REPO/adapter/authzadapter/config/authzadapter.yaml $ADAPTER_REPO/authzadapter/testdata/
cp $MIXER_REPO/template/authorization/template.yaml $ADAPTER_REPO/authzadapter/testdata/
cp $MIXER_REPO/testdata/config/attributes.yaml $ADAPTER_REPO/authzadapter/testdata/


# render the host for local test
export ADAPTER_HOST=[::]
sed -e "s/{ADAPTER_HOST}/${ADAPTER_HOST}/g" $ADAPTER_REPO/authzadapter/sample_operator_cfg.yaml > $ADAPTER_REPO/authzadapter/testdata/sample_operator_cfg.yaml

# start adapter for local test
go run $MIXER_REPO/adapter/authzadapter/cmd/main.go 45678

# start mixer server with specified config in another terminal
$GOPATH/out/linux_amd64/release/mixs server --configStoreURL=fs://$ADAPTER_REPO/authzadapter/testdata/

# run mixer client in another terminal, namespace not match, adapter should not be sent the policy check request
$GOPATH/out/linux_amd64/release/mixc check -s destination.service.host="testservice.svc.cluster.local",destination.namespace="test-namespace",request.path="/test",request.method="GET" --stringmap_attributes "request.headers=authorization:Basic dGVzdENsaWVudDpzZWNyZXQ=;x-request-priority:50"

# namesace match, adapter get the request
$GOPATH/out/linux_amd64/release/mixc check -s destination.service.host="testservice.svc.cluster.local",destination.namespace="secured-api",request.path="/test",request.method="GET" --stringmap_attributes "request.headers=authorization:Basic dGVzdENsaWVudDpzZWNyZXQ=;x-request-priority:50"

~~~

### Build Binary Execution and Docker Image

~~~ shell

# build binary
CGO_ENABLED=0 GOOS=linux go build -a -v -o $ADAPTER_REPO/authzadapter/bin/authzadapter $MIXER_REPO/adapter/authzadapter/cmd/main.go

# build docker image
docker build -t jianshao/authzadapter:0.1.0 $ADAPTER_REPO/authzadapter
# in case the build docker environment is not the same with kubernetes cluster, and you don't want to push the image to remote repository
docker save -o authzadapter.tar jianshao/authzadapter:0.1.0
# switch to the docker environment of the kubernetes cluster worker node
...
# load the image in the kubernetes cluster worker node
docker load -i authzadapter.tar

~~~

### Deploy on Kubernetes and Test

~~~ shell

# render the host for kubernetes deployment
sed -e "s/{ADAPTER_HOST}/authzadapter-service/g" $ADAPTER_REPO/authzadapter/sample_operator_cfg.yaml > $ADAPTER_REPO/authzadapter/testdata/sample_operator_cfg.yaml

# create kubernetes resources
kubectl apply -f $ADAPTER_REPO/authzadapter/authzadapter-deployment.yaml
kubectl apply -f $ADAPTER_REPO/authzadapter/testdata/template.yaml
kubectl apply -f $ADAPTER_REPO/authzadapter/testdata/authzadapter.yaml
kubectl apply -f $ADAPTER_REPO/authzadapter/testdata/sample_operator_cfg.yaml

# create test api
kubectl create ns secured-api
kubectl create ns insecure-api
kubectl label namespace secured-api istio-injection=enabled
kubectl label namespace insecure-api istio-injection=enabled

kubectl apply -f $ISTIO/istio/samples/httpbin/httpbin.yaml -n secured-api
kubectl apply -f $ISTIO/istio/samples/httpbin/httpbin.yaml -n insecure-api

# run on minikube environment
export SECURED_HTTPBIN=$(kubectl get service httpbin -n secured-api -o go-template='{{.spec.clusterIP}}')

# access secured httpbin, adapter should get the policy check request
curl -i -X POST \
-H "Authorization:Basic dGVzdENsaWVudDpzZWNyZXQ=" \
-H "Content-Type:application/json" \
-H "X-Request-Priority:50" \
-d \
'{
  "message":"hello world!"
}' \
"http://$SECURED_HTTPBIN:8000/post"

# run on minikube environment
export INSECURE_HTTPBIN=$(kubectl get service httpbin -n insecure-api -o go-template='{{.spec.clusterIP}}')

# access insecure httpbin, adapter should not get the policy check request
curl -i -X POST \
-H "Authorization:Basic dGVzdENsaWVudDpzZWNyZXQ=" \
-H "Content-Type:application/json" \
-H "X-Request-Priority:50" \
-d \
'{
  "message":"hello world!"
}' \
"http://$INSECURE_HTTPBIN:8000/post"
~~~