package pod

import (
	"bytes"
	"fmt"
	"io"

	"condenser/internal/store/psm"

	"gopkg.in/yaml.v3"
)

type PodManifest struct {
	Kind        string
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
	Containers  []psm.ContainerTemplateSpec
	Replicas    int
}

type manifestMeta struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

type podManifestSpec struct {
	Containers []containerManifest `yaml:"containers"`
}

type podManifest struct {
	APIVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   manifestMeta    `yaml:"metadata"`
	Spec       podManifestSpec `yaml:"spec"`
}

type rsTemplate struct {
	Metadata manifestMeta    `yaml:"metadata"`
	Spec     podManifestSpec `yaml:"spec"`
}

type replicaSetManifest struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   manifestMeta `yaml:"metadata"`
	Spec       struct {
		Replicas int        `yaml:"replicas"`
		Template rsTemplate `yaml:"template"`
	} `yaml:"spec"`
}

type containerManifest struct {
	Name    string           `yaml:"name"`
	Image   string           `yaml:"image"`
	Command []string         `yaml:"command"`
	Args    []string         `yaml:"args"`
	Env     []manifestEnvVar `yaml:"env"`
	Ports   []manifestPort   `yaml:"ports"`
	Tty     bool             `yaml:"tty"`
}

type manifestEnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type manifestPort struct {
	ContainerPort int `yaml:"containerPort"`
	HostPort      int `yaml:"hostPort"`
}

func DecodeK8sManifests(body []byte) ([]PodManifest, error) {
	dec := yaml.NewDecoder(bytes.NewReader(body))

	var result []PodManifest
	for {
		var raw map[string]any
		if err := dec.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(raw) == 0 {
			continue
		}
		kind, _ := raw["kind"].(string)
		if kind == "" {
			return nil, fmt.Errorf("kind is required")
		}

		rawBytes, err := yaml.Marshal(raw)
		if err != nil {
			return nil, err
		}

		switch kind {
		case "Pod":
			var pod podManifest
			if err := yaml.Unmarshal(rawBytes, &pod); err != nil {
				return nil, err
			}
			manifest := buildPodManifest(pod.Metadata, pod.Spec.Containers)
			manifest.Kind = "Pod"
			if manifest.Name == "" {
				return nil, fmt.Errorf("pod name is required")
			}
			result = append(result, manifest)
		case "ReplicaSet":
			var rs replicaSetManifest
			if err := yaml.Unmarshal(rawBytes, &rs); err != nil {
				return nil, err
			}
			meta := rs.Spec.Template.Metadata
			if meta.Name == "" {
				meta.Name = rs.Metadata.Name
			}
			manifest := buildPodManifest(meta, rs.Spec.Template.Spec.Containers)
			manifest.Kind = "ReplicaSet"
			manifest.Replicas = rs.Spec.Replicas
			if manifest.Replicas == 0 {
				manifest.Replicas = 1
			}
			if manifest.Name == "" {
				return nil, fmt.Errorf("replicaset template name is required")
			}
			result = append(result, manifest)
		default:
			return nil, fmt.Errorf("unsupported kind: %s", kind)
		}
	}

	return result, nil
}

func buildPodManifest(meta manifestMeta, containers []containerManifest) PodManifest {
	if meta.Namespace == "" {
		meta.Namespace = "default"
	}
	specs := make([]psm.ContainerTemplateSpec, 0, len(containers))
	for _, c := range containers {
		cmd := c.Command
		if len(c.Args) > 0 {
			cmd = append(append([]string{}, c.Command...), c.Args...)
		}
		envs := make([]string, 0, len(c.Env))
		for _, e := range c.Env {
			if e.Name == "" {
				continue
			}
			envs = append(envs, e.Name+"="+e.Value)
		}
		ports := make([]string, 0, len(c.Ports))
		for _, p := range c.Ports {
			if p.ContainerPort == 0 {
				continue
			}
			if p.HostPort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d", p.HostPort, p.ContainerPort))
			}
		}
		specs = append(specs, psm.ContainerTemplateSpec{
			Name:    c.Name,
			Image:   c.Image,
			Command: cmd,
			Env:     envs,
			Port:    ports,
			Tty:     c.Tty,
		})
	}
	return PodManifest{
		Name:        meta.Name,
		Namespace:   meta.Namespace,
		Labels:      meta.Labels,
		Annotations: meta.Annotations,
		Containers:  specs,
	}
}
