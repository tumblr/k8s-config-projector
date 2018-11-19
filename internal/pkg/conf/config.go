package conf

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/tumblr/k8s-config-projector/internal/pkg/version"
)

// config is the config loaded for a running instance; flags are stuffed in here!
type config struct {
	debug bool
	// manifestDir is the directory where projection manifests are loaded from
	manifestDir string
	// outputDir is where the ConfigMap yaml files will be generated in
	outputDir string
	// configDir is the root of the config repo checkout
	configDir string
	// configVersion is the label we use to identify a specific generation of configs
	configVersion   string
	labelVersionKey string
	labelManagedKey string
}

// Config is the interface for loading flag settings for the CLI app
type Config interface {
	ManifestDir() string
	OutputDir() string
	ConfigDir() string
	Debug() bool
	Version() string
	BuildDate() string
	Generation() string
	LabelVersionKey() string
	LabelManagedKey() string
}

// LoadConfigFromArgs returns a new config given some CLI args
func LoadConfigFromArgs(args []string) (Config, error) {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	c := config{}
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: (Version=%s Commit=%s Package=%s Built=%s Runtime=%s)\n", args[0], version.Version, version.Commit, version.Package, version.BuildDate, runtime.Version())
		fs.PrintDefaults()
	}

	fs.BoolVar(&c.debug, "debug", false, "Debug")
	fs.StringVar(&c.configDir, "config-repo", "", "Use this path as the root of the config directory. Projections are relative to this directory. (required)")
	fs.StringVar(&c.outputDir, "output", "", "Output generated ConfigMaps in this directory (required)")
	fs.StringVar(&c.manifestDir, "manifests", "", "Directory containing manifests yaml files (required)")
	fs.StringVar(&c.configVersion, "generation", strconv.FormatInt(time.Now().Unix(), 10), "Generation label used when annotating ConfigMaps")
	fs.StringVar(&c.labelManagedKey, "label-managed-key", "tumblr.com/managed-configmap", "Label all generated ConfigMaps with this key=true")
	fs.StringVar(&c.labelVersionKey, "label-version-key", "tumblr.com/config-version", "Label all generated ConfigMaps with this key, using the value of --generation")
	err := fs.Parse(args[1:])
	if err != nil {
		return nil, err
	}
	err = c.Validate()
	return &c, err
}

func (c *config) Validate() error {
	requiredDirs := map[string]string{
		"manifests": c.manifestDir,
		"outputDir": c.outputDir,
		"configDir": c.configDir,
	}
	for k, v := range requiredDirs {
		if v == "" {
			return fmt.Errorf("%s requires an argument", k)
		}
		if k == "outputDir" {
			if _, err := os.Stat(v); err != nil {
				return err
			}
		}
		f, err := os.Open(v)
		if err != nil {
			return err
		}
		s, err := f.Stat()
		if err != nil {
			return err
		}
		if !s.IsDir() {
			return fmt.Errorf("%s argument %s is not a directory", k, v)
		}

	}
	if c.configVersion == "" {
		return fmt.Errorf("generation argument must be specified")
	}
	return nil
}

func (c *config) Generation() string {
	return c.configVersion
}

func (c *config) ManifestDir() string {
	return c.manifestDir
}

func (c *config) OutputDir() string {
	return c.outputDir
}

func (c *config) ConfigDir() string {
	return c.configDir
}

func (c *config) Debug() bool {
	return c.debug
}

func (c *config) Version() string {
	return version.Version
}

func (c *config) BuildDate() string {
	return version.BuildDate
}

func (c *config) LabelVersionKey() string {
	return c.labelVersionKey
}

func (c *config) LabelManagedKey() string {
	return c.labelManagedKey
}
