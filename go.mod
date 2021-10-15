module github.com/ApsaraDB/PolarDB-Stack-Operator

go 1.16

require (
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/google/wire v0.5.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	github.com/ApsaraDB/PolarDB-Stack-Common v1.0.0
	github.com/ApsaraDB/PolarDB-Stack-Workflow v1.0.0
	gotest.tools/v3 v3.0.3
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/component-base v0.17.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.0
)
