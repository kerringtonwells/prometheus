# sli runner

> probes your [Concourse][concourse] installation, generating 
> [Service Level Indicators][slis] (SLIs)

![sample dashboard](https://user-images.githubusercontent.com/3574444/59943990-a8559480-9431-11e9-8a0e-0ffefa157cce.png)

[concourse]: https://concourse-ci.org 
[slis]: https://landing.google.com/sre/sre-book/chapters/service-level-objectives/

## why

Even tough Concourse emits many metrics that are useful for an operator, it
might still be hard to have a quick grasp of how high-level funcitionality is
performing.

With SLIs , one is able to better reason about what's broken on user-facing
functionality that the service exposes.

## prior

Before `slirunner`, [oxygen-mask][oxygen-mask] was the solution for running
high-level probes against Concourse installation.

It has a few quirks that I don't think are necessary to have:

- requires another Concourse installation to run those probes
- tightly coupled to datadog
- performs UI testing

[oxygen-mask]: https://github.com/concourse/oxygen-mask

## what

`slirunner` is a single [Go][go] binary that, once "installed" (run somewhere),
periodically executes several probes against Concourse, keeping track of the
successes and failures.

A consumer of `slirunner` can consume the reports from two mediums:

- [Prometheus][prometheus] exposed metrics
- structured logs

It also supports:

- single runs
- TODO: worker-related probing against multiple tags and teams

[go]: https://golang.org/
[prometheus]: https://prometheus.io/

## usage

There are several ways of running `slirunner`.

In the end, it's always the `slirunner` being run with some flags set.

```
Usage:
  slirunner [OPTIONS] <once | start>

Help Options:
  -h, --help  Show this help message

Available commands:
  once   performs a single run of the SLIs suite
  start  initiates the periodic run of the SLIs suite
```

The binary can either be built using `go`:

```console
$ go get -u github.com/cirocosta/slirunner
```

or fetched from the releases page:

```console
$ curl -SL https://github.com/cirocosta/slirunner/releases/download/v0.1.0/slirunner.tgz | tar xvzf -
fly slirunner
```

With that done, it's all about having it running:

```bash
slirunner start \
  --target $TARGET_TO_GENERATE \
  --username $CONCOURSE_BASIC_AUTH_USERNAME \
  --password $CONCOURSE_BASIC_AUTH_PASSWORD \
  --concourse-url $URL_OF_THE_CONCOURSE_INSTALLATION
```

### using docker / kubernetes

A container image `cirocosta/slirunner` is continuously pushed to
https://hub.docker.com/r/cirocosta/slirunner.

e.g., using `docker-compose`:

```yaml
version: '3'
services:
  slirunner:
    image: cirocosta/slirunner
    command:
      - start
      - --target=test
      - --concourse-url=http://web:8080
      - --password=test
      - --username=test
```

For kubernetes, check out the example under
[`./examples/kubernetes.yaml`](./examples/kubernetes.yaml).

## license

See [./LICENSE](./LICENSE)
