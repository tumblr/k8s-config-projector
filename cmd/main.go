package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tumblr/k8s-config-projector/internal/pkg/conf"
	"github.com/tumblr/k8s-config-projector/internal/pkg/version"
	"github.com/tumblr/k8s-config-projector/pkg/output"
	"github.com/tumblr/k8s-config-projector/pkg/types/v1/manifest"
)

const (
	// ConfigMapSizeLimit is the max size of a ConfigMap we support (in string format).
	// This is pretty rough, but it is intended to avoid the 1M etcd max
	// There are a few limits:
	// * ConfigMaps cannot be over 1M.  This is etcd's limitation.
	// * Internally, `kubectl apply` creates annotations, which has a size limit of 256K. This translates to a ConfigMap that can't be over ~512K.
	// For reference, these are size limits discussions:
	// * https://github.com/coreos/prometheus-operator/issues/535#issuecomment-319659063
	// * https://github.com/kubernetes/kubernetes/issues/15878#issuecomment-149728026
	ConfigMapSizeLimit = 500000 // 500k max!
)

//
// this program assumes the config repos are cloned into local directories
// the repo-to-directory mappings are read from repos yaml file, which is passed by cli
// the manifest directory with the user defined config projection is also passed as cli arg
func main() {
	// CLI
	c, err := conf.LoadConfigFromArgs(os.Args)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
	log.Printf("Starting up. version=%s commit=%s branch=%s built=%s runtime=%s", version.Version, version.Commit, version.Branch, version.BuildDate, runtime.Version())
	if c.Debug() {
		log.Printf("config base path: %s\n", c.ConfigDir())
		log.Printf("manifest directory: %s\n", c.ManifestDir())
		log.Printf("output directory: %s\n", c.OutputDir())
	}

	manifests, err := loadManifests(c)
	if err != nil {
		log.Fatal("error loading projection manifests: " + err.Error())
	}

	if len(manifests) == 0 {
		log.Fatal("No manifest loaded! Aborting\n")
	}

	// timestamp
	tUnix := time.Now().Unix()

	// project each config file into a separate ConfigMap
	for _, m := range manifests {
		fname := filepath.Join(c.OutputDir(), output.BuildFileOutputName(m.GetNamespace(), m.GetName(), tUnix))
		log.Printf("Writing ConfigMap %s/%s to %s", m.GetNamespace(), m.GetName(), fname)
		cfgString, err := m.ProjectConfigMapAsYAML()
		if err != nil {
			log.Fatalf("unable to project %s/%s: %s", m.GetNamespace(), m.GetName(), err.Error())
		}

		// before we write this out, lets make sure the byte size isnt exceeding our hardcoded limit
		if len([]byte(cfgString)) > ConfigMapSizeLimit {
			log.Fatalf("generated ConfigMap for %s/%s that was %d bytes, exceeding size limit of %d bytes\nYou may want to split this projection into multiple ConfigMaps to reduce size", m.GetNamespace(), m.GetName(), len([]byte(cfgString)), ConfigMapSizeLimit)
		}

		err = ioutil.WriteFile(fname, []byte(cfgString), 0600)
		if err != nil {
			log.Fatalf("unable to write config to %s: %s", fname, err.Error())
		}

	}
}

// loadManifests takes a conf.Config, determines the root manifest path, and recursively finds all
// yaml manifests under it. It loads, parses, and validates the manifests before returning them in
// a map from "namespace/name" -> manifest.ConfigProjectionManifest
func loadManifests(cfg conf.Config) (manifests map[string]manifest.ConfigProjectionManifest, err error) {
	rootpath := cfg.ManifestDir()
	err = filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.New(path + ": " + err.Error())
		}
		m, err := manifest.LoadFromYAMLBytes(raw, cfg)
		if err != nil {
			return errors.New(path + ": " + err.Error())
		}
		key := fmt.Sprintf("%s/%s", m.Namespace, m.Name)
		if _, ok := manifests[key]; ok {
			return fmt.Errorf("duplicate projection mapping found at namespace=%s name=%s file=%s", m.Namespace, m.Name, path)
		}
		manifests[key] = m
		return nil
	})
	return manifests, err
}
