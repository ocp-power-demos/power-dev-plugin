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

#### Debug DaemonSet
To debug the running plugin, you can use: 

```
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

Thse are commented out in the DaemonSet.

#### Debug Kubelet

You can check the kubelet behavior using:

```
# journalctl -u kubelet
...
 7446 handler.go:95] "Registered client" name="power-dev-plugin/dev"
wrapper[7446]: I1219 04:32:20.722778    7446 manager.go:230] "Device plugin connected" resourceName="power-dev-plugin/dev"
wrapper[7446]: I1219 04:32:20.723559    7446 client.go:93] "State pushed for device plugin" resource="power-dev-plugin/dev" re>
wrapper[7446]: I1219 04:32:20.726284    7446 manager.go:279] "Processed device updates for resource" resourceName="power-dev-p>
wrapper[7446]: I1219 04:32:27.293908    7446 setters.go:333] "Updated capacity for device plugin" plugin="power-dev-plugin/dev>
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