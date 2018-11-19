# Projection Manifests

## Schema

A projection manifest has a name, namespace, and a list of [datasources](/docs/datasource.md).

```yaml
---
name: "config-projection-name-here"
namespace: "namespace-for-configmap"
data: [] # list of datasources
```

## Examples

```yaml
---
name: notifications-us-east-1-production
namespace: notification-production
data:
# extract some fields from JSON
- source: generated/us-east-1/production/config.json
  output_file: config.json
  field_extraction:
  - memcached_hosts: $.memcached.notifications.production.hosts
  - flags: $.applications.notification.production.flags
  - log_level: $.applications.notification.production.log_level
# extract a scalar value from a YAML
- source: apps/us-east-1/production/notification.yaml
  output_file: launch_flags
  extract: $.launch_flags
```

## Datasources

See [datasources](/docs/datasource.md) for more documentation on how each datasource can be configured.
