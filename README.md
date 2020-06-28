# hazelcast-go-operator

A Kubernetes Operator based on the Operator SDK for creating and managing Hazelcast instances.

## Current status

The project is kind of **toy project** for now so it is at **very early alpha** stage.

## Installation

You can find latest images at [DockerHub](https://hub.docker.com/repository/docker/hasancelik/hazelcast-go-operator).

After create Kubernetes cluster at your local or one of the cloud environments, you can play the operator via below commands:

`kubectl apply --recursive -f deploy/crds/hazelcast.com_hazelcasts_crd.yaml`

`kubectl apply --recursive -f deploy/operator-rbac.yaml`

`kubectl apply --recursive -f deploy/operator.yaml`

```
kubectl get pods

NAME                                     READY   STATUS    RESTARTS   AGE
hazelcast-go-operator-8646c77fb8-rj922   1/1     Running   0          3s

kubectl logs -f hazelcast-go-operator-8646c77fb8-rj922

{"level":"info","ts":1577659310.4765203,"logger":"cmd","msg":"Operator Version: 0.0.1"}
{"level":"info","ts":1577659310.476789,"logger":"cmd","msg":"Go Version: go1.13.5"}
{"level":"info","ts":1577659310.476876,"logger":"cmd","msg":"Go OS/Arch: linux/amd64"}
{"level":"info","ts":1577659310.4769561,"logger":"cmd","msg":"Version of operator-sdk: v0.13.0"}
{"level":"info","ts":1577659310.477187,"logger":"leader","msg":"Trying to become the leader."}
{"level":"info","ts":1577659310.6937504,"logger":"leader","msg":"No pre-existing lock was found."}
{"level":"info","ts":1577659310.6981168,"logger":"leader","msg":"Became the leader."}
{"level":"info","ts":1577659310.9068968,"logger":"controller-runtime.metrics","msg":"metrics server is starting to listen","addr":"0.0.0.0:8383"}
{"level":"info","ts":1577659310.9072833,"logger":"cmd","msg":"Registering Components."}
{"level":"info","ts":1577659311.3391876,"logger":"metrics","msg":"Metrics Service object created","Service.Name":"hazelcast-go-operator-metrics","Service.Namespace":"default"}
{"level":"info","ts":1577659311.542264,"logger":"cmd","msg":"Could not create ServiceMonitor object","error":"no ServiceMonitor registered with the API"}
{"level":"info","ts":1577659311.5423012,"logger":"cmd","msg":"Install prometheus-operator in your cluster to create ServiceMonitor objects","error":"no ServiceMonitor registered with the API"}
{"level":"info","ts":1577659311.5423844,"logger":"cmd","msg":"Starting the Cmd."}
{"level":"info","ts":1577659311.542574,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"hazelcast-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1577659311.5427518,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"hazelcast-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1577659311.5428214,"logger":"controller-runtime.controller","msg":"Starting Controller","controller":"hazelcast-controller"}
{"level":"info","ts":1577659311.5429094,"logger":"controller-runtime.manager","msg":"starting metrics server","path":"/metrics"}
{"level":"info","ts":1577659311.6431444,"logger":"controller-runtime.controller","msg":"Starting workers","controller":"hazelcast-controller","worker count":1}
```

`kubectl apply --recursive -f deploy/hazelcast-rbac.yaml`

`kubectl apply --recursive -f deploy/crds/hazelcast.com_v1alpha1_hazelcast_cr.yaml`

```
kubectl get pods

NAME                                     READY   STATUS    RESTARTS   AGE
example-hazelcast-0                      1/1     Running   0          28s
example-hazelcast-1                      1/1     Running   0          26s
example-hazelcast-2                      1/1     Running   0          24s
hazelcast-go-operator-8646c77fb8-rj922   1/1     Running   0          2m59s

kubectl logs -f example-hazelcast-0
########################################
# JAVA_OPTS=-Dhazelcast.mancenter.enabled=false -Djava.net.preferIPv4Stack=true -Djava.util.logging.config.file=/opt/hazelcast/logging.properties -Dhazelcast.rest.enabled=true -Dhazelcast.config=/data/hazelcast/hazelcast.yaml
# CLASSPATH=/opt/hazelcast/*:/opt/hazelcast/lib/*
# starting now....
########################################
+ exec java -server -Dhazelcast.mancenter.enabled=false -Djava.net.preferIPv4Stack=true -Djava.util.logging.config.file=/opt/hazelcast/logging.properties -Dhazelcast.rest.enabled=true -Dhazelcast.config=/data/hazelcast/hazelcast.yaml com.hazelcast.core.server.StartServer
Dec 29, 2019 10:44:21 PM com.hazelcast.config.AbstractConfigLocator
INFO: Loading configuration '/data/hazelcast/hazelcast.yaml' from System property 'hazelcast.config'
Dec 29, 2019 10:44:21 PM com.hazelcast.config.AbstractConfigLocator
INFO: Using configuration file at /data/hazelcast/hazelcast.yaml
Dec 29, 2019 10:44:22 PM com.hazelcast.instance.AddressPicker
INFO: [LOCAL] [dev] [3.12.5] Prefer IPv4 stack is true, prefer IPv6 addresses is false
Dec 29, 2019 10:44:22 PM com.hazelcast.instance.AddressPicker
INFO: [LOCAL] [dev] [3.12.5] Picked [10.24.0.57]:5701, using socket ServerSocket[addr=/0.0.0.0,localport=5701], bind any local is true
Dec 29, 2019 10:44:22 PM com.hazelcast.system
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Hazelcast 3.12.5 (20191210 - 294ff46) starting at [10.24.0.57]:5701
Dec 29, 2019 10:44:22 PM com.hazelcast.system
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Copyright (c) 2008-2019, Hazelcast, Inc. All Rights Reserved.
Dec 29, 2019 10:44:22 PM com.hazelcast.spi.impl.operationservice.impl.BackpressureRegulator
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Backpressure is disabled
Dec 29, 2019 10:44:22 PM com.hazelcast.internal.config.ConfigValidator
WARNING: Property hazelcast.rest.enabled is deprecated. Use configuration object/element instead.
Dec 29, 2019 10:44:23 PM com.hazelcast.spi.discovery.integration.DiscoveryService
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Kubernetes Discovery properties: { service-dns: null, service-dns-timeout: 5, service-name: hazelcast-service, service-port: 0, service-label: null, service-label-value: true, namespace: default, pod-label: null, pod-label-value: null, resolve-not-ready-addresses: false, use-node-name-as-external-address: false, kubernetes-api-retries: 3, kubernetes-master: https://kubernetes.default.svc}
Dec 29, 2019 10:44:23 PM com.hazelcast.spi.discovery.integration.DiscoveryService
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Kubernetes Discovery activated with mode: KUBERNETES_API
Dec 29, 2019 10:44:23 PM com.hazelcast.instance.Node
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Activating Discovery SPI Joiner
Dec 29, 2019 10:44:23 PM com.hazelcast.spi.impl.operationexecutor.impl.OperationExecutorImpl
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Starting 2 partition threads and 3 generic threads (1 dedicated for priority tasks)
Dec 29, 2019 10:44:23 PM com.hazelcast.internal.diagnostics.Diagnostics
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Diagnostics disabled. To enable add -Dhazelcast.diagnostics.enabled=true to the JVM arguments.
Dec 29, 2019 10:44:23 PM com.hazelcast.core.LifecycleService
INFO: [10.24.0.57]:5701 [dev] [3.12.5] [10.24.0.57]:5701 is STARTING
Dec 29, 2019 10:44:26 PM com.hazelcast.spi.discovery.integration.DiscoveryService
WARNING: [10.24.0.57]:5701 [dev] [3.12.5] Cannot fetch the current zone, ZONE_AWARE feature is disabled
Dec 29, 2019 10:44:26 PM com.hazelcast.kubernetes.KubernetesClient
WARNING: Cannot fetch public IPs of Hazelcast Member PODs, you won't be able to use Hazelcast Smart Client from outside of the Kubernetes network
Dec 29, 2019 10:44:26 PM com.hazelcast.nio.tcp.TcpIpConnector
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Connecting to /10.24.2.20:5701, timeout: 10000, bind-any: true
Dec 29, 2019 10:44:26 PM com.hazelcast.nio.tcp.TcpIpConnection
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Initialized new cluster connection between /10.24.0.57:44945 and /10.24.2.20:5701
Dec 29, 2019 10:44:30 PM com.hazelcast.nio.tcp.TcpIpConnection
INFO: [10.24.0.57]:5701 [dev] [3.12.5] Initialized new cluster connection between /10.24.0.57:5701 and /10.24.0.58:48357
Dec 29, 2019 10:44:31 PM com.hazelcast.internal.cluster.ClusterService
INFO: [10.24.0.57]:5701 [dev] [3.12.5]

Members {size:1, ver:1} [
	Member [10.24.0.57]:5701 - 0dcd9277-2f4e-402e-9b59-54f4844df9f1 this
]

Dec 29, 2019 10:44:31 PM com.hazelcast.core.LifecycleService
INFO: [10.24.0.57]:5701 [dev] [3.12.5] [10.24.0.57]:5701 is STARTED
Dec 29, 2019 10:44:37 PM com.hazelcast.internal.cluster.ClusterService
INFO: [10.24.0.57]:5701 [dev] [3.12.5]

Members {size:3, ver:3} [
	Member [10.24.0.57]:5701 - 0dcd9277-2f4e-402e-9b59-54f4844df9f1 this
	Member [10.24.0.58]:5701 - 5f8c4179-39c5-4623-9bf3-83b160fe871a
	Member [10.24.2.20]:5701 - 7247ac61-b164-4afa-a79f-7d08f9c16792
]
```
