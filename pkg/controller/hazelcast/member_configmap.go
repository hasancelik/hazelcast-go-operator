package hazelcast

import (
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-go-operator/pkg/apis/hazelcast/v1alpha1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type MemberConfig struct {
	Hazelcast struct {
		Network struct {
			Join struct {
				Kubernetes struct {
					Enabled     bool   `yaml:"enabled"`
					ServiceName string `yaml:"service-name"`
					Namespace   string `yaml:"namespace"`
				}
			}
		}
	}
}

func createDefaultMemberConfig(cr *hazelcastv1alpha1.Hazelcast) (string, error){
	defaultConfigDataTemplate := []byte(`
hazelcast:
 network:
  join:
   kubernetes:
    enabled:
    service-name:
    namespace:
`)
	memberConfig := MemberConfig{}
	err := yaml.Unmarshal(defaultConfigDataTemplate, &memberConfig)
	if err != nil {
		return "", err
	}
	kubernetes := &memberConfig.Hazelcast.Network.Join.Kubernetes
	kubernetes.Enabled = true
	kubernetes.ServiceName = HazelcastServiceName
	kubernetes.Namespace = cr.Namespace

	defaultConfig, err := yaml.Marshal(&memberConfig)
	if err != nil {
		return "", err
	}
	return string(defaultConfig), nil
}

func configDataFromSpec(cr *hazelcastv1alpha1.Hazelcast) (string, error){
	if cr.Spec.Config.Data == nil {
		defaultMemberConfig, err := createDefaultMemberConfig(cr)
		if err != nil {
			return "", err
		}
		return defaultMemberConfig, nil
	} else {
		return cr.Spec.Config.Data[HazelcastConfigFileName], nil
	}
}

func GetHazelcastConfigMap(cr *hazelcastv1alpha1.Hazelcast) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	configMap.ObjectMeta = metav1.ObjectMeta{
		Name:      HazelcastConfigName,
		Namespace: cr.Namespace,
	}

	configYAMLData, err := configDataFromSpec(cr)
	if err != nil {
		return nil, err
	}
	configMap.Data = map[string]string{}
	configMap.Data[HazelcastConfigFileName] = configYAMLData

	hash := generateSHA1CheckSum(&configYAMLData)
	configMap.Annotations = map[string]string{
		"lastConfigHash": hash,
	}

	return configMap, nil
}

func (r *ReconcileHazelcast) configMapForHazelcast(cr *hazelcastv1alpha1.Hazelcast) *v1.ConfigMap {
	configMap, err := GetHazelcastConfigMap(cr)
	if err != nil {
		log.Error(err, "error updating hazelcast config")
		return nil
	}

	controllerutil.SetControllerReference(cr, configMap, r.scheme)
	return configMap
}

func GetHazelcastConfigMapSelector(cr *hazelcastv1alpha1.Hazelcast) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      HazelcastConfigName,
	}
}
