# Power Device Plugin

Power Device Plugin to add protected devices into a non-privileged container.

The Power Device Plugin uses the [Kubernetes Device Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) in order to add specific devices to the given Pod.

## Steps

### Device Plugin
The device plugin only installs on the workers.

1. To deploy the device plugin: 

``` shell
# kustomize build manifests | oc apply -f -

project.project.openshift.io/power-device-plugin created
serviceaccount/power-device-plugin created
clusterrolebinding.rbac.authorization.k8s.io/power-device-plugin created
daemonset.apps/power-device-plugin created
```

### Sample

1. To deploy the sample: `kustomize build examples | oc apply -f -`

## Sources

1. https://github.com/intel/intel-device-plugins-for-kubernetes/blob/main/pkg/deviceplugin/manager.go#L96
2. https://github.com/kairen/simple-device-plugin/tree/master
3. https://github.com/kubernetes/kubelet/tree/master/pkg/apis/deviceplugin/v1beta1
4. https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/#examples

## Build

The build includes multiple architectures: `linux/amd64`, `linux/ppc64le`, `linux/s390x`.
The build uses the [ubi9/ubi:9.4](https://catalog.redhat.com/software/containers/ubi9/ubi/615bcf606feffc5384e8452e?architecture=ppc64le&image=676258d7607921b4d7fcb8c8&gti-tabs=unauthenticated) image.