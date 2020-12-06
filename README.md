# drone-tardigrade

drone-ci [plugin](https://docs.drone.io/plugins/overview/) for pushing artifacts to the Tardigrade network

## Build
Build the container with the following commands
```
docker build -t kristaxox/drone-tardigrade .
```

## Usage
Execute from the working directory:
```
docker run --rm \
    -e DRY_RUN=true \
    -e PLUGIN_ACCESS=<tardigrade-access> \
    -e PLUGIN_BUCKET=<bucket> \
    -e PLUGIN_SOURCE=<source> \
    -e PLUGIN_TARGET=<target> \
    -v $(pwd):$(pwd) \
    -w $(pwd) \
    kristaxox/drone-tardigrade
```