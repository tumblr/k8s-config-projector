package manifest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/tumblr/k8s-config-projector/internal/pkg/conf"
	_ "github.com/tumblr/k8s-config-projector/internal/pkg/testing"
)

var (
	TestConfigBasePath          = "test/sources"
	ExpectedTestProjectionsPath = "test/fixtures" // /namespace/name/<files>
	ManifestsPath               = "test/manifests"
	testConfig, _               = conf.LoadConfigFromArgs([]string{
		"-debug=false",
		"-output=test/",
		"-manifests=" + ManifestsPath,
		"-generation=unittest123",
		("-config-repo=" + TestConfigBasePath),
	})
	cfg conf.Config = testConfig
	// all the test manifests to load
	testManifests, _    = filepath.Glob(fmt.Sprintf("%s/*.yaml", ManifestsPath))
	parseErrorManifests = map[string]string{
		"test/manifests/parseerrors/1.yaml": "source files of glob format cannot specify a `output_file` field for projection",
		"test/manifests/parseerrors/2.yaml": "unsupported output format; must be one of raw, yaml, or json",
		"test/manifests/parseerrors/3.yaml": "unsupported source format; must be file, glob, yaml, or json",
		"test/manifests/parseerrors/4.yaml": "you cannot use this output format without either `extract` or `field_extractions`",
		"test/manifests/parseerrors/5.yaml": "output_file field required for this projection type",
		"test/manifests/parseerrors/6.yaml": "absolute paths for `source` are not permitted",
		"test/manifests/parseerrors/7.yaml": "name must only consist of lower case alphanumeric characters, -, and . and be 253 chars or less",
		"test/manifests/parseerrors/8.yaml": "namespace must only consist of lower case alphanumeric characters, -, and . and be 253 chars or less",
	}
)

func TestLoadManifestFromFile(t *testing.T) {
	for _, f := range testManifests {
		t.Logf("loading %s\n", f)
		c, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		m, err := LoadFromYAMLBytes(c, cfg)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range m.Data {
			t.Logf("%s/%s: %s\n", m.Namespace, m.Name, d.String())
		}
	}
}

func TestLoadManifestFromFileWithErrors(t *testing.T) {
	for f, errString := range parseErrorManifests {
		t.Logf("loading %s\n", f)
		c, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		_, err = LoadFromYAMLBytes(c, cfg)
		if err == nil {
			t.Fatalf("Expected %s error, but got nothing", errString)
		}
		if err.Error() != errString {
			t.Fatalf("Expected %s error, but got %s", errString, err.Error())
		}
	}
}

func TestLoadManifestFromFileAndProject(t *testing.T) {
	for _, f := range testManifests {
		t.Logf("loading %s\n", f)
		c, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		m, err := LoadFromYAMLBytes(c, cfg)
		if err != nil {
			t.Fatal(err)
		}

		// now, project this manifest
		t.Logf("loaded %s; projecting with basepath %s\n", m.String(), cfg.ConfigDir())
		cm, err := m.Project()
		if err != nil {
			t.Errorf("Unable to project %s/%s with config basepath: %s\n", m.Namespace, m.Name, cfg.ConfigDir())
			t.Fatal(err)
		}

		// make sure the right number of fixtures are present as data items
		expectedDataPath := fmt.Sprintf("%s/%s/%s", ExpectedTestProjectionsPath, m.Namespace, m.Name)
		t.Logf("loading expected fixtures from %s\n", expectedDataPath)
		expectedCMFiles, err := ioutil.ReadDir(expectedDataPath)
		if err != nil {
			t.Fatal(err)
		}
		if len(cm.Data) != len(expectedCMFiles) {
			t.Fatalf("Expected %d data items in config map %s/%s, but got %d", len(expectedCMFiles), m.Namespace, m.Name, len(cm.Data))
		}
		// now, read in each test fixture output file and assert they are the same
		for _, finfo := range expectedCMFiles {
			x, ok := cm.Data[finfo.Name()]
			if !ok {
				t.Fatalf("expected to find item in configmap %s/%s '%s', but did not", m.Namespace, m.Name, finfo.Name())
			}
			expected, err := ioutil.ReadFile(path.Join(expectedDataPath, finfo.Name()))
			if err != nil {
				t.Fatal(err)
			}
			// NOTE: the file we read from disk always has the trailing EOF, so lets rip that thing off just for
			// clarity.
			expected = bytes.TrimSuffix(expected, []byte("\n"))
			if bytes.Compare(expected, []byte(x)) != 0 {
				// NOTE: diff.CharacterDiff is expected, actual
				// so (~~X~~) means expected has X while actual does not
				// and (++X++) means actual has X while expected does not
				t.Fatalf("expected configmap %s/%s item '%s' to not be different from expected fixture:\n%s\ndiff:%s", m.Namespace, m.Name, finfo.Name(), x, diff.CharacterDiff(string(expected), x))
			}
		}
		t.Logf("OK! %s was projected successfully!\n", f)
	}
}
