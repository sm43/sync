module knative.dev/sample-controller

go 1.15

require (
	github.com/go-logr/zapr v0.2.0
	github.com/manifestival/client-go-client v0.5.0
	github.com/manifestival/manifestival v0.6.0
	github.com/tektoncd/pipeline v0.28.0
	go.uber.org/zap v1.19.1
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	k8s.io/api v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v0.21.4
	k8s.io/code-generator v0.21.4
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a // indirect
	knative.dev/hack v0.0.0-20211222071919-abd085fc43de
	knative.dev/hack/schema v0.0.0-20211222071919-abd085fc43de
	knative.dev/pkg v0.0.0-20211206113427-18589ac7627e
)
