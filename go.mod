module github.com/jianshaow/istio-mixer-adapter

go 1.13

require (
	github.com/gogo/googleapis v1.1.0
	github.com/gogo/protobuf v1.3.0
	google.golang.org/grpc v1.24.0
	istio.io/api v0.0.0-20191115173247-e1a1952e5b81
	istio.io/gogo-genproto v0.0.0-20191029161641-f7d19ec0141d
	istio.io/istio v0.0.0-20191202224512-41dee99277db
	istio.io/pkg v0.0.0-20191030005435-10d06b6b315e
)

replace k8s.io/api => k8s.io/api v0.0.0-20191003000013-35e20aa79eb8

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191003002041-49e3d608220c

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191003002408-6e42c232ac7d

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191003002707-f6b7b0f55cc0
