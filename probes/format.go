package probes

import (
	"bytes"
	"text/template"
	"fmt"

	"github.com/pkg/errors"
)

type Config struct {
	Username string
	Password string
	LdapAuth bool
	LdapTeam string
	WorkerPool string
	Harbor_url string
	ConcourseUrl string
	InsecureTls  bool
	Target           string
	ExistingPipeline string
	Pipeline         string
}

func FormatProbe(formatting string, c Config) (res string) {
	fmt.Println("FormatProbe")
	fmt.Println(formatting)
	var buf = new(bytes.Buffer)

	tmpl, err := template.New("").Parse(formatting)
	if err != nil {
		panic(errors.Wrapf(err, "failed to parse template"))
	}

	err = tmpl.Execute(buf, c)
	if err != nil {
		panic(errors.Wrapf(err, "failed to execute template"))
	}

	res = buf.String()
	return
}
