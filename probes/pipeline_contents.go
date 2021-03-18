package probes

const pipelineContents = `
resources:
- name: time-trigger
  type: time
  source: {interval: 24h}
  tags:
  - ((workerpool_name))

jobs:
- name: simple-job
  build_logs_to_retain: 20
  public: true
  plan:
  - &say-hello
    task: say-hello
    config:
      platform: linux
      image_resource:
        type: registry-image
        source: {repository: busybox}
      run:
        path: echo
        args: ["Hello, world!"]
    tags:
    - ((workerpool_name))

- name: failing
  build_logs_to_retain: 20
  public: true
  plan:
  - task: fail
    config:
      platform: linux
      image_resource:
        type: registry-image
        source: {repository: busybox}
      run:
        path: /bin/false
    tags:
    - ((workerpool_name))

- name: auto-triggering
  build_logs_to_retain: 20
  public: true
  plan:
  - get: time-trigger
    trigger: true
    tags:
    - ((workpool_name))
  - *say-hello
`
