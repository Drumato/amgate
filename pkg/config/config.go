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
	Matchers []MatcherConfig   `yaml:"matchers"`
	Name     string            `yaml:"name"`
	Attrs    map[string]string `yaml:"attrs,omitempty"`
}

type MatcherConfig struct {
	Key   string `yaml:"key"`
	Op    string `yaml:"op"`
	Value string `yaml:"value"`

	Labels            LabelMatcherConfig `yaml:"labels,omitempty"`
	Annotations       LabelMatcherConfig `yaml:"annotations,omitempty"`
	CommonLabels      LabelMatcherConfig `yaml:"commonLabels,omitempty"`
	CommonAnnotations LabelMatcherConfig `yaml:"commonAnnotations,omitempty"`
}

type LabelMatcherConfig struct {
	Matchers []MatcherConfig `yaml:"matchers"`
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

func (c *Config) ValidateAndDefault() error {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	for i := range c.Actions {
		if c.Actions[i].Name == "" {
			return errors.New("action name is required")
		}

		for j := range c.Actions[i].Matchers {
			if err := c.Actions[i].Matchers[j].ValidateAndDefault(); err != nil {
				return err
			}
		}
		if c.Actions[i].Attrs == nil {
			c.Actions[i].Attrs = map[string]string{}
		}
	}

	return nil
}

func (m *MatcherConfig) ValidateAndDefault() error {
	if m.Key == "" {
		return errors.New("matcher key is required")
	}
	if m.Op == "" {
		return errors.New("matcher op is required")
	}
	if m.Op != "=" && m.Op != "!=" && m.Op != "=~" {
		return errors.New("matcher op must be = or != or =~")
	}
	if m.Value == "" {
		return errors.New("matcher value is required")
	}

	if m.Labels.Matchers == nil {
		m.Labels.Matchers = []MatcherConfig{}
	}
	if m.Annotations.Matchers == nil {
		m.Annotations.Matchers = []MatcherConfig{}
	}
	if m.CommonLabels.Matchers == nil {
		m.CommonLabels.Matchers = []MatcherConfig{}
	}
	if m.CommonAnnotations.Matchers == nil {
		m.CommonAnnotations.Matchers = []MatcherConfig{}
	}

	for _, subM := range m.Labels.Matchers {
		if err := subM.ValidateAndDefault(); err != nil {
			return err
		}
	}
	for _, subM := range m.Annotations.Matchers {
		if err := subM.ValidateAndDefault(); err != nil {
			return err
		}
	}
	for _, subM := range m.CommonLabels.Matchers {
		if err := subM.ValidateAndDefault(); err != nil {
			return err
		}
	}
	for _, subM := range m.CommonAnnotations.Matchers {
		if err := subM.ValidateAndDefault(); err != nil {
			return err
		}
	}

	return nil
}
