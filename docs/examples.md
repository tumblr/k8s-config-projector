# Examples

Here are some full examples of projection manifests, along with some source data.

## Structured Field Extractions

Source config file `above/file.json`:

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

Projection manifest yaml:

```yaml
---
name: example-config
namespace: my-namespace
data:
- source: above/file.json
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
