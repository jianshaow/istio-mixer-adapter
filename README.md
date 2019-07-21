# istio-mixer-authz-adapter

First of all, get go environment ready.

~~~ shell
mkdir -p $GOPATH/src/istio.io/ && \
cd $GOPATH/src/istio.io/  && \
git clone https://github.com/istio/istio

export ISTIO=$GOPATH/src/istio.io
export MIXER_REPO=$GOPATH/src/istio.io/istio/mixer

pushd $ISTIO/istio && make mixs
pushd $ISTIO/istio && make mixc

cp authzadapter $MIXER_REPO/adapter/ -r
cd $MIXER_REPO/adapter/authzadapter

go generate ./...
go build ./...

cp config/authzadapter.yaml testdata/

go run cmd/main.go 45678

$GOPATH/out/linux_amd64/release/mixs server --configStoreURL=fs://${MIXER_REPO}/adapter/authzadapter/testdata

$GOPATH/out/linux_amd64/release/mixc check -s destination.service="svc.cluster.local" --stringmap_attributes "request.headers=Authorization:Basic c2VsZlNlcnZpY2VBUFA6c2VjcmV0"
~~~