# k8s-config-projector

Turn configuration files on disk into structured Kubernetes [ConfigMaps](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#create-a-configmap)

![GitHub release](https://img.shields.io/github/release/tumblr/k8s-config-projector.svg) ![Travis (.org)](https://img.shields.io/travis/tumblr/k8s-config-projector.svg) ![Docker Automated build](https://img.shields.io/docker/automated/tumblr/k8s-config-projector.svg) ![Docker Build Status](https://img.shields.io/docker/build/tumblr/k8s-config-projector.svg) ![MicroBadger Size](https://img.shields.io/microbadger/image-size/tumblr/k8s-config-projector.svg) ![Docker Pulls](https://img.shields.io/docker/pulls/tumblr/k8s-config-projector.svg) ![Docker Stars](https://img.shields.io/docker/stars/tumblr/k8s-config-projector.svg)  [![Godoc](https://godoc.org/github.com/tumblr/k8s-config-projector?status.svg)](http://godoc.org/github.com/tumblr/k8s-config-projector)

# What is this?

At Tumblr, we have many sources for logical configuration data that defines how deployed services operate. Some of these sources are rapidly changing, some are updated automatically, and some are changed by humans. Because we always strive to limit complexity in our stack, we wanted to come up with a way that applications can be made aware of these changes as they happen, without becoming tightly coupled to our internal processes. Additionally, because there are many configuration sources (git repos, bots, cronjobs, tools like [Collins](https://tumblr.github.io/collins/), etc), and we wanted to write tools to manipulate this configuration data after being generated, we generate intermediary representations in a `config` git repo. However, because each application has very unique needs for what data it requires to run (eg TLS certs, comma separated lists of host names, env vars, specifically formatted JSON or YAML), we needed a method to transform the common configuration files into formats each application could consume or further process, without hardcoding configuration into our images, or requiring a rebuild/redeploy of the app each time configuration data changes. To accomplish this task, we built the `k8s-config-projector`.

## In a nutshell

The k8s-config-projector is a program to project config files from disk into Kubernetes `ConfigMap`s through Continuous Integration (CI). The `ConfigMap`s can be applied to a k8s cluster with `kubectl apply -f <filename>` (or `create`+`replace`) during Continuous Delivery (CD). These `ConfigMap`s can then be consumed by applications in Kubernetes, by mounting them as volumes or environment variables inside pods.

This is quite useful to inform applications in Kubernetes about changing configuration. For example, this can be used to discover the nodes participating in a memcached ring, configure an application with feature flags, etc.

To use this tool, it is expected that:

1. You have some configuration data (YAML/JSON/raw files) you want to get into your pods without baking them into the Docker images
2. That configuration data comes from some files on disk (we call this the "config repo").
3. The files are either binary blobs, or YAML/JSON structured data
4. You have some CI/CD mechanism you want to hook up to, so ConfigMaps are automatically generated and updated when you merge changes to your `config` repo

If you want to use this to manage secrets, you should check out [k8s-secret-projector](https://github.com/tumblr/k8s-secret-projector) instead!

```
+--------------------------+
|                          |
| Projection Mappings Repo +-------------+ YAML Projection
| (--manifests)            |             | Manifests
|                          |             |
+--------------------------+             |
                                         |
                                         |
+-----------------+                  +---v--------------------+                       +------------------+
|                 |  config sources  |                        |                       |                  |
| Config Repo     +------------------>                        | generates configmaps  | Output Directory |
| (--config-repo) |                  |  k8s-config-projector  +-----------------------> (--output)       |
|                 |                  |                        |                       |                  |
+-----------------+                  |                        |                       +---------+--------+
                                     +------------------------+                                 |
                                                                                                |
                                                                                                |
                                                                                                | creates ConfigMaps
                                                        +----------------------+                | in Kubernetes
                                                        |                      |                |
                                                        |  kubectl create -f   <----------------+
                                                        |                      |
                                                        +----------------------+

```

## Functionality

The projector can:

* Take raw files and stuff them into a ConfigMap
* Glob files in your config repo, and stuff ALL of them in your configmap
* Extract fields from your structured data (yaml/json)
* Create new structured outputs from a subset of a yaml/json source by pulling out some fields and dropping others
* Translate back and forth between JSON and YAML (convert a YAML source to a JSON output, etc)
* Support for extracting complex fields like objects+arrays from sources, and not just scalars!

# Documentation

You should get started by checking out some example projection manifests.

* Projection Manifests documentation: [projection_manifests.md](/docs/projection_manifests.md)
* `DataSource` schema documentation: [datasource.md](/docs/datasource.md)

# Hacking

Check out [hacking.md](/docs/hacking.md) for how to build this!

# Run

## Requirements

1. Have your `--config-repo` checked out, and location in `${CONFIG_REPO}`
2. Have your path to your projection mappings (`--manifests`) in `${MANIFESTS_REPO}`
3. Create an `${OUTPUT_DIR}` to keep the generated manifests the tool spits out

## Running with Docker

```shell
$ docker run -it --rm \
    -v "${OUTPUT_DIR}:/output" \
    -v "${MANIFESTS_REPO}:/manifests:ro" \
    -v "${CONFIG_REPO}:/config:ro" \
    tumblr/k8s-config-projector:latest
```

Collect your generated ConfigMaps in `${OUTPUT_DIR}`!

## Running from binary

```shell
$ make && ./bin/k8s-config-projector --debug --manifests=${MANIFESTS_REPO} --config-repo=${CONFIG_REPO}  --output=${OUTPUT_DIR}
...
```

Collect your generated ConfigMaps in `${OUTPUT_DIR}`!

# How to use ConfigMap in a pod

Example config map:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: test-config-1
  namespace: foo-bar
data:
  test_config.txt: |
    this is a config file with a whole bunch of stuff
    to show that configs can be any form
    blah
    blah
    blah

```

Example pod config:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod1
  namespace: mynamespace
spec:
  containers:
    - name: busybox-test
      image: k8s.gcr.io/busybox
      command: [ "/bin/sh", "-c", "cat /etc/config/test-config.txt" ]
      volumeMounts:
      - name: config-volume
        mountPath: /etc/config
  volumes:
    - name: config-volume
      configMap:
        name: test-config-1
        items:
        - key: test_config.txt
          path: test-config.txt
  restartPolicy: Never
```

# Limitations

There are a few things to note when using the Config Projector. They are actually limits of Kubernetes/`etcd`, but it still is useful to be aware of them.

* ConfigMaps cannot be over 1M.  Because ConfigMaps are stored in Kubernetes API, the ConfigMaps are backed by `etcd`. This has a 1M limit of each object stored.
* Internally, `kubectl apply` creates annotations, which has a size limit of 256K. This translates to a ConfigMap that can't be over ~512K. If you want to avoid this, use `kubectl create` instead.

For reference, these are size limits discussions:

* https://github.com/coreos/prometheus-operator/issues/535#issuecomment-319659063
* https://github.com/kubernetes/kubernetes/issues/15878#issuecomment-149728026

# License

[Apache 2.0](/LICENSE.txt)

Copyright 2018, Tumblr, Inc.
