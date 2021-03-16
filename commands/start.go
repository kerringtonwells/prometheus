package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerctx"
	"github.com/cirocosta/slirunner/exporter"
	"github.com/cirocosta/slirunner/probes"
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

	Prometheus exporter.Exporter `group:"Prometheus configuration"`
}

func (c *startCommand) Execute(args []string) (err error) {
	var (
		allProbes = probes.NewAll(
			c.Target,
			c.Username, c.Password,
			c.ConcourseUrl,
			c.PipelinesPrefix,
			c.InsecureTls,
			c.LdapAuth, c.LdapTeam,
		)
		ticker = time.NewTicker(c.Interval)
	)

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
