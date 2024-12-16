# Power Device Plugin
Power Device Plugin to add protected devices into a non-privileged container.

The Power Device Plugin uses the [Kubernetes Device Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) in order to add specific devices to the given Pod.

https://github.com/kubernetes/kubelet/tree/master/pkg/apis/deviceplugin/v1beta1
https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/#examples


1. To deploy the sample: `kustomize build examples | oc apply -f -`