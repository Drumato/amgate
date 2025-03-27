package config

import (
	"context"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Config represents the entire configuration of amgate.
// that is stored in ConfigMap.
type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Actions []ActionConfig `yaml:"actions"`
}

// ServerConfig represents the configuration of the server.
type ServerConfig struct {
	// Host is the host of the server.
	Host string `yaml:"host"`
	// Port is the port of the server.
	Port int `yaml:"port"`
}

// ActionConfig represents the configuration of an action.
type ActionConfig struct {
	Matchers []MatcherConfig `yaml:"matchers"`
	Actor    ActorConfig     `yaml:"actor"`
}

type MatcherConfig struct {
	Key   string `yaml:"key"`
	Op    string `yaml:"op"`
	Value string `yaml:"value"`

	Labels []LabelMatcherConfig `yaml:"labels,omitempty"`
}

type LabelMatcherConfig struct {
	Matchers []MatcherConfig `yaml:"matchers"`
}

type ActorConfig struct {
	Kind string `yaml:"kind"`
	// TODO: rollout strategy
	CustomAttrs map[string]string `yaml:"custom_attrs,omitempty"`
}

func LoadFromConfigMap(
	ctx context.Context,
	k8sClient client.Client,
) (Config, error) {
	cmNamespace := os.Getenv("AMGATE_NAMESPACE")
	cmNamespace = lo.If(cmNamespace != "", cmNamespace).Else("amgate-system")

	cmName := os.Getenv("AMGATE_CONFIGMAP_NAME")
	cmName = lo.If(cmName != "", cmName).Else("amgate-config")

	cm := corev1.ConfigMap{}
	if err := k8sClient.Get(ctx, types.NamespacedName{
		Namespace: cmNamespace,
		Name:      cmName,
	}, &cm); err != nil {
		return Config{}, errors.WithStack(err)
	}

	cfg := Config{}
	if v, ok := cm.Data["server"]; ok {
		sc := ServerConfig{}
		if err := yaml.Unmarshal([]byte(v), &sc); err != nil {
			return Config{}, errors.WithStack(err)
		}
		cfg.Server = sc
	}

	if v, ok := cm.Data["actions"]; ok {
		ac := []ActionConfig{}
		if err := yaml.Unmarshal([]byte(v), &ac); err != nil {
			return Config{}, errors.WithStack(err)
		}
		cfg.Actions = ac
	}

	return cfg, nil
}
