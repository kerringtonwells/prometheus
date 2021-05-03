package commands

import (
	"context"
	"fmt"
	"os"
	"time"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerctx"
	"github.com/kerringtonwells/slirunner/exporter"
	"github.com/kerringtonwells/slirunner/probes"
	"strings"
)

type startCommand struct {
	Target          string        `long:"target"   required:"true"       description:"target to be used by fly commands"`
	PipelinesPrefix string        `long:"prefix"   default:"slirunner-"  description:"prefix used in pipelines created by probes"`
	Interval        time.Duration `long:"interval" default:"1m"          description:"interval between executions"`

	Username     string `long:"username"      short:"u" required:"true"  description:"username of a local user"`
	Password     string `long:"password"      short:"p" required:"true"  description:"password of the local user"`
	ConcourseUrl string `long:"concourse-url" short:"c" required:"true"  description:"URL of the concourse to monitor"`
	InsecureTls  bool   `long:"insecure-tls"  short:"k" required:"false" description:"Skip tls verification"`

	LdapAuth bool   `long:"ldapauth"      short:"l" required:"false" description:"LDAP boolean if using ldap auth"`
	LdapTeam string `long:"ldapteam"      short:"m" required:"false" description:"LDAP team if using ldap auth"`
	WorkerPool string `long:"workerpool"      short:"w" required:"true" description:"worker pool for concourse pipelines"`
	Harbor_url string `long:"harbor_url"      short:"r" required:"true" description:"repository url"`
	Prometheus exporter.Exporter `group:"Prometheus configuration"`
	Debug string `long:"debug"      short:"d" required:"true" description:"debug"`
}
//This is the first thin that gets run. Its goes into all.go and uses probes.NewAll too get allProbes
func (c *startCommand) Execute(args []string) (err error) {
	var logs string
	if strings.Contains(c.Debug, "true") {
	    logs = "set -o xtrace"
	}else {
      logs = ""
	}

	//singleConcourseUrl := strings.Split(c.ConcourseUrl, " ")
	//singleTarget := strings.Split(c.Target, " ")
	//singleWorkerPool := strings.Split(c.WorkerPool, " ")
	//counter := 0
	//var stringCounter string
  //fmt.Println(singleConcourseUrl)
	//for i := range singleConcourseUrl {
	var (
		// Passing the variables from the struct above to the NewAll function in all.go
		allProbes = probes.NewAll(
			c.Target,
			c.Username,
			c.Password,
			c.ConcourseUrl,
			c.PipelinesPrefix,
			c.InsecureTls,
			c.LdapAuth,
			c.LdapTeam,
			c.WorkerPool,
			c.Harbor_url,
			logs,
		)
		ticker = time.NewTicker(c.Interval)
	)
	//}

	ctx, cancel := context.WithCancel(context.Background())
	go onTerminationSignal(func() {
		cancel()
		c.Prometheus.Close()
	})

	logger := lager.NewLogger("slirunner").Session("run")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	ctx = lagerctx.NewContext(ctx, logger)

	go func() {
		err := c.Prometheus.Listen()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}

		cancel()
	}()

	allProbes.Run(ctx)
	fmt.Println("Probes ")
  fmt.Println(ctx)
	for {
		select {
		case <-ticker.C:
			allProbes.Run(ctx)
		case <-ctx.Done():
			c.Prometheus.Close()
			return
		}
	}
	return
}
