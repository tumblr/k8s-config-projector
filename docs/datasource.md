# Datasources

A Projection Mapping Datasource is something that defines how to turn a source file (configuration file) into a projected output file in a given format. Datasources can be "projected" into a set of data that will be used to build a `ConfigMap`. Each datasource can result in 1 or more projected items in the ConfigMap.

NOTE: a Datasource is part of the larger schema for [projection manifests](/docs/projection_manifests.md). A datasource only makes sense in a projection mapping file.

## Schema

A datasource has the following schema:

```yaml
source: "some/file.json"
output_file: "myconfig.json"
source_format: file|glob|yaml|json
# note: only one of extract, field_extractions may be used
extract: "$.json.path[2].notation.scalar"
field_extractions:
- hostname: "$.data.hostname"
- rack_position: "$.data.rack_position"
output_format: raw|json|yaml
```

### Source

A source is a relative path to the `--config-repo` argument. The file here is loaded at projection time, and used to extract fields with either `extract`, `field_extractions`, or injected wholesale into the `ConfigMap` when `output_format: raw`.

It is useful to use structured (yaml or json) sources, to enable surgical field extraction. This allows downstream consumers of the produced `ConfigMap` to take a least-surface-area approach to configuration. Alternatively, files can be injected wholesale with `output_format: raw`.

### Source Types

Supported sources are:

* `file`: This is a raw, unstructured file. You cannot extract fields from this source. This is default if you omit both `extract` and `field_extractions`.
* `json`: Enables structured field extraction. This is inferred if source ends in `.json`.
* `yaml`: Enables structured field extraction. This is inferred if source ends in `.yaml`.
* `glob`: If source contains `*`, this is assumed. No structured field extraction capability, but this allows you to project multiple files into your ConfigMap

### Output File

This is what the ConfigMap's data is named. This will be used as the projected filename for a ConfigMap when used as a Volume in Kubernetes.

Its useful to make this field's name reflect the type of data contained. I.e. if you are doing `field_extractions`, you should name this so it reflects the `output_format`

### Extract

This uses `jsonpath` notation to pull out a scalar value from the `source`, and make its value the data of this `DataSource` when projected. See https://github.com/oliveagle/jsonpath for docs on the syntax.

Example source:

```yaml
foo:
  bar:
  - hello
  - world
nodes:
- hostname: foo-12345.domain.tld
  ip: 1.2.3.4
- hostname: bar-56849.domain.tld
  ip: 2.3.4.5
```

Example field extraction for `foo-12345`: `$.nodes[0].hostname`

### Field Extractions

This is the same as above, but with the ability to pull out multiple fields and create a new structured output (yaml or json) with only some keys. Keeping the above source, we can create a new YAML output containing only fields we care about

```yaml
source: above/file.json
output_format: yaml
output_file: something.yaml
field_extractions:
  # lets pull out "world" as scope
  scope: $.foo.bar[1]
  # this can extract and project whole structured subsets! note that we dont pull out both fields, but the whole map here
  node: $.nodes[0]
```

Resulting configmap:

```yaml
kind: ConfigMap
apiVersion: v1
metadata
  name: something
  namespace: some-namespace
  labels:
    tumblr.com/config-version: "1539778254"
    tumblr.com/managed-configmap: "true"
  data:
    something.yaml: |
      scope: "world"
      node:
        hostname: foo-12345.domain.tld
        ip: 1.2.3.4
```

### Output File

This determines the key in `data:` in the ConfigMap. This will be the "filename" your configmap's data elements project onto the filesystem as. This should match your output_format suffix.

### Output Format

* `raw`: just the unadulterated file
* `json`: inferred when output_file is `*.json`. Will format structured data in JSON.
* `yaml`: inferred when output_file is `*.yaml`. Will format structured data in YAML.


