# Hacking

## Build

* `make` builds the binary
* `make docker` builds the docker container

## Dependencies

To download the dependencies, typically this happens as part of `make`, with `make vendor`.

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
