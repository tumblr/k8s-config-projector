package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/tumblr/k8s-config-projector/internal/pkg/conf"
	"github.com/tumblr/k8s-config-projector/pkg/types"
	ds "github.com/tumblr/k8s-config-projector/pkg/types/v1/datasource"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/printers"
)

var (
	nameValidationRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9\-\.]+$`)
)

// ConfigProjectionManifest is the user-supplied config ConfigProjectionManifest
type ConfigProjectionManifest struct {
	Name      string           `yaml:"name"`
	Namespace string           `yaml:"namespace"`
	Data      []*ds.DataSource `yaml:"data"`

	c conf.Config
}

// GetName - return the name of the ConfigProjectionManifest
func (m *ConfigProjectionManifest) GetName() string {
	return m.Name
}

// GetNamespace - return the k8s namespace
func (m *ConfigProjectionManifest) GetNamespace() string {
	return m.Namespace
}

// String returns a string rep for debugging
func (m *ConfigProjectionManifest) String() string {
	items := []string{}
	for _, x := range m.Data {
		items = append(items, x.String())
	}
	sort.Strings(items)

	return fmt.Sprintf("%s/%s(%s)", m.Namespace, m.Name, strings.Join(items, ","))
}

// Project - return the config projections of the ConfigProjectionManifest
// AsConfigMap - returns a ProjectionMapping projected into a ConfigMap with all fields
// extracted and transformed into the k8s resource
// https://v1-7.docs.kubernetes.io/docs/api-reference/v1.7/#configmap-v1-core
// https://godoc.org/k8s.io/api/core/v1#ConfigMap
func (m *ConfigProjectionManifest) Project() (v1.ConfigMap, error) {
	basePath := m.c.ConfigDir()

	// each []byte is a projected file, each key is a file name
	dataList := map[string]string{}
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels: map[string]string{
				m.c.LabelVersionKey(): m.c.Generation(),
				m.c.LabelManagedKey(): "true",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
	}

	for _, d := range m.Data {
		projectedDataItems, err := d.Project(basePath)
		if err != nil {
			return cm, err
		}
		for k, v := range projectedDataItems {
			if _, ok := dataList[k]; ok {
				return cm, errors.New("duplicate projection key " + k + " in projection sources")
			}
			// NOTE: the ConfigMap takes strings, not []bytes so we need to type conversions
			// to bring []byte into strings
			dataList[k] = string(v)
		}
	}
	cm.Data = dataList
	return cm, nil
}

// SetDefaults after loading from a yaml
func (m *ConfigProjectionManifest) SetDefaults() error {
	for _, d := range m.Data {
		err := d.SetDefaults()
		if err != nil {
			return err
		}
	}
	return nil
}

// ProjectConfigMapAsYAML projects the manifest as a yaml marshalled string
func (m *ConfigProjectionManifest) ProjectConfigMapAsYAML() (string, error) {
	cm, err := m.Project()
	if err != nil {
		return "", err
	}
	printer := printers.YAMLPrinter{}
	buf := bytes.NewBuffer([]byte{})
	err = printer.PrintObj(&cm, buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// LoadFromYAMLBytes - load a ConfigProjectionManifest from a byte slice typically read from IO
func LoadFromYAMLBytes(raw []byte, cfg conf.Config) (ConfigProjectionManifest, error) {
	var m ConfigProjectionManifest
	err := yaml.UnmarshalStrict(raw, &m)
	if err != nil {
		return m, err
	}
	m.SetDefaults()
	m.c = cfg
	err = m.Validate()
	if err != nil {
		return m, err
	}
	return m, nil
}

// Validate validates a ProjectionManifest
func (m *ConfigProjectionManifest) Validate() error {
	if m.Name == "" {
		return types.ErrMissingName
	}
	if m.Namespace == "" {
		return types.ErrMissingNamespace
	}
	// validate name and namespace meets k8s requirements:
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	if len(m.Name) > 253 || !nameValidationRegexp.MatchString(m.Name) {
		return types.ErrInvalidName
	}
	if len(m.Namespace) > 253 || !nameValidationRegexp.MatchString(m.Namespace) {
		return types.ErrInvalidNamespace
	}
	for _, d := range m.Data {
		err := d.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
