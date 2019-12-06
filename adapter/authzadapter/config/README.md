# Authorization Adapter Config

## Generation Protobuf Code

The protobuf code is already generated and commit to repository, you don't need to do this again. The following just shows how to do this.

~~~ shell

# checkout the adapter source code
cd /tmp && \
   git clone https://github.com/jianshaow/istio-mixer-adapter.git

export GOPATH=$HOME/go
export ADAPTER_REPO=/tmp/istio-mixer-adapter

mkdir -p $GOPATH/src/istio.io
cd $GOPATH/src/istio.io
git clone https://github.com/istio/istio

export ISTIO=$GOPATH/src/istio.io/istio

cd $ISTIO
# base on stable version
git checkout 1.2.9

cp $ADAPTER_REPO/adapter/authzadapter $ISTIO/mixer/adapter/

go generate $ISTIO/mixer/adapter/authzadapter/

~~~
