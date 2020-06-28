package hazelcast

const (
	HazelcastImage                         = "hazelcast/hazelcast"
	HazelcastVersion                       = "4.0"
	HazelcastStatefulSetName               = "hazelcast"
	HazelcastServiceAccountName            = "hazelcast-service-account"
	HazelcastServiceName                   = "hazelcast-service"
	HazelcastConfigName                    = "hazelcast-config"
	HazelcastConfigFileName                = "hazelcast.yaml"
	HazelcastPort                   int32  = 5701
	HazelcastPortName                      = "hazelcast"
	HazelcastSecurityContextUser    int64  = 65534
	HazelcastLivenessProbeEndpoint  string = "/hazelcast/health/node-state"
	HazelcastReadinessProbeEndpoint string = "/hazelcast/health/ready"
)
