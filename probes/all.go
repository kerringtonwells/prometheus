package probes

import (
	"github.com/kerringtonwells/slirunner/runnable"
	"os"
	"time"
)

func NewLogin(target string, username string, password string, concourseUrl string, insecureTls bool, ldapAuth bool, ldapTeam string, workerpool string) runnable.Runnable {
	var (
		config = Config{
			Target:       target,
			Username:     username,
			Password:     password,
			ConcourseUrl: concourseUrl,
			InsecureTls:  insecureTls,
			LdapAuth:     ldapAuth,
			LdapTeam:     ldapTeam,
			WorkerPool:   workerpool,
		}
		timeout = 60 * time.Second
	)

	var loginCommand string

  if ldapAuth {
		loginCommand = `

        CONCOURSE_URL="{{ .ConcourseUrl }}"
        CONCOURSE_USER="{{ .Username }}"
        CONCOURSE_PASSWORD="{{ .Password }}"
        CONCOURSE_TARGET="{{ .Target }}"
        > token
        LDAP_AUTH_URL=$CONCOURSE_URL$(curl -k -b token -c token -L "$CONCOURSE_URL/sky/login" -s | grep "/sky/issuer/auth" | awk -F'"' '{print $4}')
        curl -k -s -b token -c token -L --data-urlencode "login=$CONCOURSE_USER" --data-urlencode "password=$CONCOURSE_PASSWORD" "$LDAP_AUTH_URL"
        ATC_BEARER_TOKEN=$(cat token | grep skymarshal_auth0  | cut -f 7 | tr -d \" | sed 's/bearer//')

        cat <<ENDOFSCRIPT > ~/.flyrc
targets:
  $CONCOURSE_TARGET:
    api: $CONCOURSE_URL
    insecure: true
    team: {{ .LdapTeam }}
    token:
      type: Bearer
      value: $ATC_BEARER_TOKEN
ENDOFSCRIPT

        fly -t {{ .Target }} status

        `
} else if insecureTls {
	loginCommand = `
	fly -t {{ .Target }} login -u {{ .Username }} -p {{ .Password }} -c {{ .ConcourseUrl }} -k
	`
} else {
	loginCommand = `
	fly -t {{ .Target }} login -u {{ .Username }} -p {{ .Password }} -c {{ .ConcourseUrl }}
	`
}

	return runnable.NewWithLogging("login",
		runnable.NewWithMetrics("login",
			runnable.NewWithTimeout(
				runnable.NewShellCommand(FormatProbe(loginCommand, config), os.Stderr),
				timeout,
			),
		),
	)
}

func NewSync(target string) runnable.Runnable {
	var (
		config  = Config{Target: target}
		timeout = 4 * time.Minute
	)

	return runnable.NewWithLogging("sync",
		runnable.NewWithMetrics("sync",
			runnable.NewWithTimeout(
				runnable.NewShellCommand(FormatProbe(`

	fly -t {{ .Target }} sync

				`, config), os.Stderr),
				timeout,
			),
		),
	)
}

func NewCreateAndRunNewPipeline(target, prefix string, pipelineContentsWorker string, workerpool string, harbor_url string, logs string) runnable.Runnable {
	var (
		config = Config{
			Target:   target,
			Pipeline: prefix + "create-and-run-new-pipeline",
		}
		timeout = 180 * time.Second
	)
	return runnable.NewWithLogging("create-and-run-new-pipeline",
		runnable.NewWithMetrics("create-and-run-new-pipeline",
			runnable.NewWithTimeout(
				runnable.NewShellCommand(FormatProbe(`

set -o errexit
`+logs+`

echo "Printing the log level and the target"
echo `+target+`
echo "====================="
echo "Yaml with workerpool"
echo '`+ pipelineContentsWorker +`'
echo "====================="

fly -t {{ .Target }} destroy-pipeline -n -p {{ .Pipeline }} || true
fly -t {{ .Target }} set-pipeline -n -p {{ .Pipeline }} -c <(echo '`+pipelineContentsWorker+`')
fly -t {{ .Target }} unpause-pipeline -p {{ .Pipeline }}

wait_for_build () {
	fly -t {{ .Target }} builds -j {{ .Pipeline }}/auto-triggering | \
		grep -v pending | \
		wc -l
}

until [ "$(wait_for_build)" -gt 0 ]; do
	echo 'waiting for job to automatically trigger...'
	sleep 1
done

fly -t {{ .Target }} watch -j {{ .Pipeline }}/auto-triggering
fly -t {{ .Target }} destroy-pipeline -n -p {{ .Pipeline }}

				`, config), os.Stderr),
				timeout,
			),
		),
	)
}

func NewHijackFailingBuild(target, prefix string, pipelineContentsWorker string, logs string) runnable.Runnable {
	var (
		config = Config{
			Target:   target,
			Pipeline: prefix + "hijack-failing-build",
		}
		timeout = 60 * time.Second
	)

	return runnable.NewWithLogging("hijack-failing-build",
		runnable.NewWithMetrics("hijack-failing-build",
		  runnable.NewWithTimeout(
				runnable.NewShellCommand(FormatProbe(`

	set -o errexit
  `+logs+`
  echo "Printing the log level and the target"
	echo `+target+`
  echo "====================="
  echo "Yaml with workerpool"
	echo '`+ pipelineContentsWorker +`'
  echo "====================="
	fly -t {{ .Target }} set-pipeline -n -p {{ .Pipeline }} -c <(echo '`+pipelineContentsWorker+`')
	fly -t {{ .Target }} unpause-pipeline -p {{ .Pipeline }}
	job_name={{ .Pipeline }}/failing
	fly -t {{ .Target }} trigger-job -j "$job_name" -w || true
	build=$(fly -t {{ .Target }} builds -j "$job_name" | head -1 | awk '{print $3}')
	fly -t {{ .Target }} hijack -j "$job_name" -b $build echo Hello World
				`, config), os.Stderr),
				timeout,
			),
		),
	)
}

func NewRunExistingPipeline(target, prefix string, pipelineContentsWorker string, logs string) runnable.Runnable {
	var (
		config = Config{
			Target:   target,
			Pipeline: prefix + "run-existing-pipeline",
		}
		timeout = 60 * time.Second
	)

	return runnable.NewWithLogging("run-existing-pipeline",
		runnable.NewWithMetrics("run-existing-pipeline",
			runnable.NewWithTimeout(
				runnable.NewShellCommand(FormatProbe(`

	set -o errexit
	`+logs+`
  echo "Printing the log level and the target"
	echo `+target+`
	echo "====================="
  echo "Yaml with workerpool"
	echo '`+ pipelineContentsWorker +`'
  echo "====================="
	fly -t {{ .Target }} set-pipeline -n -p {{ .Pipeline }} -c <(echo '`+pipelineContentsWorker+`')
	fly -t {{ .Target }} unpause-pipeline -p {{ .Pipeline }}
	fly -t {{ .Target }} trigger-job -w -j "{{ .Pipeline }}/simple-job"
	echo "=========TARGET STATUS"
	fly -t `+target+` status
	echo "=========TARGET STATUS"
	fly logout -a

				`, config), os.Stderr),
				timeout,
			),
		),
	)
}


func NewAll(target []string, username string, password string, concourseUrl []string, prefix string, insecureTls bool, ldapAuth bool, ldapTeam string, workerpool []string, harbor_url string, logs string) runnable.Runnable {
  if ldapAuth && len(ldapTeam) == 0 {
			  // Assigning ldapTeam to a default team if none is given
    ldapTeam = "concourse-monitoring"
  }

  var runable_output runnable.Runnable
	firstRun := true
	for i := range concourseUrl {
    pipelineContentsWorker := `
resources:
- name: time-trigger
  type: time
  source: {interval: 24h}
  tags:
  - `+workerpool[i]+`
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
        source: {repository: ` + harbor_url + `/busybox}
      run:
        path: echo
        args: ["Hello, world!"]
    tags:
    - `+workerpool[i]+`
- name: failing
  build_logs_to_retain: 20
  public: true
  plan:
  - task: fail
    config:
      platform: linux
      image_resource:
        type: registry-image
        source: {repository: ` + harbor_url + `/busybox}
      run:
        path: /bin/false
    tags:
    - `+workerpool[i]+`
- name: auto-triggering
  build_logs_to_retain: 20
  public: true
  plan:
  - get: time-trigger
    trigger: true
    tags:
    - `+workerpool[i]+`
  - *say-hello
`
		runable_output = runnable.NewSequentially([]runnable.Runnable{
			NewLogin(target[i], username, password, concourseUrl[i], insecureTls, ldapAuth, ldapTeam, workerpool[i]),
			NewSync(target[i]),
			runnable.NewConcurrently([]runnable.Runnable{
				NewCreateAndRunNewPipeline(target[i], prefix, pipelineContentsWorker, workerpool[i], harbor_url, logs),
				NewHijackFailingBuild(target[i], prefix, pipelineContentsWorker, logs),
				NewRunExistingPipeline(target[i], prefix, pipelineContentsWorker, logs),
			}),
		})
	}
	return runable_output
}
