# istio-mixer-authz-adapter

First of all, get go environment ready.

~~~ shell
cd /tmp
git clone https://github.com/JianshaoWu/istio-mixer-authz-adapter.git

mkdir -p $GOPATH/src/istio.io/ && \
cd $GOPATH/src/istio.io/  && \
git clone https://github.com/istio/istio

export ISTIO=$GOPATH/src/istio.io
export MIXER_REPO=$GOPATH/src/istio.io/istio/mixer

pushd $ISTIO/istio && make mixs
pushd $ISTIO/istio && make mixc

cp /tmp/istio-mixer-authz-adapter/authzadapter $MIXER_REPO/adapter/ -r
cd /tmp/istio-mixer-authz-adapter/

go generate $MIXER_REPO/adapter/authzadapter/
# go generate needs docker, may need root privilege, if so, GOPATH needs to be passed for sudo as following
# sudo GOPATH=$GOPATH go generate $MIXER_REPO/adapter/authzadapter/
go build $MIXER_REPO/adapter/authzadapter/

cp $MIXER_REPO/adapter/authzadapter/config/authzadapter.yaml $MIXER_REPO/adapter/authzadapter/testdata/

export ADDRESS=[::]:45678
sed -e "s/{ADDRESS}/${ADDRESS}/g" /tmp/istio-mixer-authz-adapter/sample_operator_cfg.yaml > $MIXER_REPO/adapter/authzadapter/testdata/sample_operator_cfg.yaml

go run $MIXER_REPO/adapter/authzadapter/cmd/main.go 45678

$GOPATH/out/linux_amd64/release/mixs server --configStoreURL=fs://${MIXER_REPO}/adapter/authzadapter/testdata

$GOPATH/out/linux_amd64/release/mixc check -s destination.service.host="testservice.svc.cluster.local",destination.namespace="test-namespace",request.path="/test",request.method="GET" --stringmap_attributes "request.headers=authorization:Basic dGVzdENsaWVudDpzZWNyZXQ=;x-request-priority:50"

# real kubernetes cluster deployment

cd /tmp/istio-mixer-authz-adapter
CGO_ENABLED=0 GOOS=linux go build -a -v -o bin/authzadapter $MIXER_REPO/adapter/authzadapter/cmd/main.go

docker build -t mymixeradapter/authzadapter:1.0 .
# in case the build docker environment is not the same with kubernetes cluster, and you don't want to push the image to remote repository
docker save -o authzadapter.tar mymixeradapter/authzadapter:1.0
# switch to the docker environment of the kubernetes cluster worker node
...
# load the image in the kubernetes cluster worker node
docker load -i authzadapter.tar

sed -e "s/{ADDRESS}/authzadapter-service/g" sample_operator_cfg.yaml > authzadapter/testdata/sample_operator_cfg.yaml

kubectl apply -f authzadapter-deployment.yaml
kubectl apply -f authzadapter/testdata/template.yaml
kubectl apply -f $MIXER_REPO/adapter/authzadapter/config/authzadapter.yaml
kubectl apply -f authzadapter/testdata/sample_operator_cfg.yaml

~~~