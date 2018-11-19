package datasource

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/tumblr/k8s-config-projector/pkg/types"
)

// DataSource represents a config file source, resulting in 1 or more projected files
// Files can be raw, or extract fields from structured json/yaml
type DataSource struct {
	Source     string `yaml:"source"`
	OutputFile string `yaml:"output_file,omitempty"`
	// Format is the source format
	SourceFormat SourceFormat `yaml:"source_format,omitempty"`
	Extract      string       `yaml:"extract,omitempty"`
	// FieldExtractions are the fields to extract from structured sources
	FieldExtractions map[string]string `yaml:"field_extractions,omitempty"`
	OutputFormat     OutputType        `yaml:"output_format,omitempty"`
}

// SourceFormat is a type of input format
type SourceFormat string

// OutputType is a type of output format
type OutputType string

const (
	// FormatFile is the default input format
	FormatFile SourceFormat = "file"
	// FormatGlob takes a glob and expands to multiple raw outputs
	FormatGlob SourceFormat = "glob"
	// FormatYAML reads a file in as yaml structured data. requires extract/field_extractions
	FormatYAML SourceFormat = "yaml"
	// FormatJSON reads a file in as json structured data. requires extract/field_extractions
	FormatJSON SourceFormat = "json"

	// OutputRaw outputs normal files
	OutputRaw OutputType = "raw"
	// OutputJSON outputs extracted fields as JSON subsets (requires field_extractions)
	OutputJSON OutputType = "json"
	// OutputYAML outputs extracted fields as YAML subsets (requires field_extractions)
	OutputYAML OutputType = "yaml"
)

// isGlobSource tells us if the DataSource uses globs (not one file)
func (f *DataSource) isGlobSource() bool {
	return strings.Contains(f.Source, `*`)
}

// if source format is empty, infers proper format, or errors
func (f *DataSource) inferredSourceFormat() (SourceFormat, error) {
	if f.isGlobSource() {
		return FormatGlob, nil
	}
	if f.Source != "" {
		if f.Extract == "" && len(f.FieldExtractions) == 0 {
			return FormatFile, nil
		} else if (f.Extract != "" || len(f.FieldExtractions) > 0) && strings.HasSuffix(f.Source, ".json") {
			return FormatJSON, nil
		} else if (f.Extract != "" || len(f.FieldExtractions) > 0) && strings.HasSuffix(f.Source, ".yaml") {
			return FormatYAML, nil
		}
	}
	return "", types.ErrUnableToInferSourceFormat
}

// if output format is empty, infers proper format, or errors
func (f *DataSource) inferredOutputFormat() (OutputType, error) {
	// if we are doing field extraction and source is json and no output format specified, assume json
	if strings.HasSuffix(f.Source, ".json") && len(f.FieldExtractions) > 0 {
		return OutputJSON, nil
	}
	// if we are doing field extraction and source is yaml and no output format specified, assume yaml
	if strings.HasSuffix(f.Source, ".yaml") && len(f.FieldExtractions) > 0 {
		return OutputYAML, nil
	}
	if f.SourceFormat == FormatGlob {
		return OutputRaw, nil
	}
	if (f.SourceFormat == FormatFile && f.Extract == "" && len(f.FieldExtractions) == 0) ||
		(f.SourceFormat == FormatJSON && f.Extract != "") ||
		(f.SourceFormat == FormatYAML && f.Extract != "") {
		return OutputRaw, nil
	}
	return "", types.ErrUnableToInferOutputFormat
}

// SetDefaults after loading from a yaml
func (f *DataSource) SetDefaults() error {
	if f.SourceFormat == "" {
		sf, err := f.inferredSourceFormat()
		if err != nil {
			return err
		}
		f.SourceFormat = sf
	}

	if f.OutputFormat == "" {
		of, err := f.inferredOutputFormat()
		if err != nil {
			return err
		}
		f.OutputFormat = of
	}

	if f.OutputFile == "" && f.OutputFormat == OutputRaw && f.SourceFormat == FormatFile {
		// assume the OutputFile is the same name as the source, without any directory component!
		f.OutputFile = path.Base(f.Source)
	}
	return nil
}

// Project will take a base path and project the source into a list of byte arrays
// performing any extraction and globbing necessary
func (f *DataSource) Project(basePath string) (map[string][]byte, error) {
	projectedFiles := map[string][]byte{}

	switch f.SourceFormat {
	case FormatGlob:
		files, err := filepath.Glob(filepath.Join(basePath, f.Source))
		if err != nil {
			return nil, err
		}
		for _, v := range files {
			// extract the filename without any paths
			relativeSource := strings.TrimLeft(strings.Split(v, basePath)[1], "/")
			name := relativeSource[strings.LastIndex(relativeSource, "/")+1:]

			// check for duplicate files in the output bucket before reading files
			if _, ok := projectedFiles[name]; ok {
				// already exists a projection with this name. abort!
				return nil, errors.New("existing file projection with name " + name)
			}

			source := path.Join(basePath, relativeSource)
			buf, err := ioutil.ReadFile(source)
			if err != nil {
				return nil, err
			}
			// because we are globbing files from the filesystem, remove the trailing \n always
			// TODO(gabe) i dunno if this is appropriate; we really need to strip the trailing non-printing
			// char that is always present when we read from disk?
			projectedFiles[name] = bytes.TrimSuffix(buf, []byte("\n"))
		}
	case FormatFile:
		// its just a single raw file extraction, read from Source and return its contents
		if _, ok := projectedFiles[f.OutputFile]; ok {
			return nil, errors.New("existing file projection with name " + f.OutputFile)
		}

		source := path.Join(basePath, f.Source)
		buf, err := ioutil.ReadFile(source)
		if err != nil {
			return nil, err
		}
		projectedFiles[f.OutputFile] = bytes.TrimSuffix(buf, []byte("\n"))
	case FormatJSON:
		return f.projectJSON(basePath)
	case FormatYAML:
		return f.projectYAML(basePath)
	default:
		return nil, types.ErrUnsupportedSourceType
	}
	return projectedFiles, nil
}

// String returns a string of the DataSource
func (f *DataSource) String() string {
	return fmt.Sprintf("DataSource{%s:%s} output=%s extract=%s fields=%s", f.Source, f.SourceFormat, f.OutputFormat, f.Extract, f.FieldExtractions)
}

// Validate validates a DataSource
func (f *DataSource) Validate() error {
	if f.SourceFormat != FormatFile && f.SourceFormat != FormatGlob && f.SourceFormat != FormatJSON && f.SourceFormat != FormatYAML {
		return types.ErrUnsupportedSourceFormat
	}
	if f.OutputFormat != OutputJSON && f.OutputFormat != OutputYAML && f.OutputFormat != OutputRaw {
		return types.ErrUnsupportedOutputFormat
	}
	if f.isGlobSource() && f.OutputFormat != OutputRaw {
		return types.ErrSourceGlobWithRawOutput
	}
	if f.OutputFile == "" && (f.OutputFormat == OutputRaw || f.OutputFormat == OutputJSON || f.OutputFormat == OutputYAML) && f.SourceFormat != FormatGlob {
		return types.ErrOutputFileRequired
	}
	if f.Extract == "" && len(f.FieldExtractions) == 0 && f.OutputFormat != OutputRaw {
		return types.ErrOutputFormatRequiresExtractors
	}
	if f.OutputFormat == OutputRaw && len(f.FieldExtractions) != 0 {
		return types.ErrWrongOutputFormatWithFieldExtractions
	}
	if f.OutputFile != "" && f.SourceFormat == FormatGlob {
		return types.ErrFormatGlobRequiresNoOutputFile
	}
	if path.IsAbs(f.Source) {
		return types.ErrAbsolutePathSource
	}
	return nil
}
