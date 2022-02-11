package polaris

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/mapping"
)

type target struct {
	Addr      string        `key:",optional"`
	Service   string        `key:",optional"`
	Namespace string        `key:"namespace,optional"`
	Timeout   time.Duration `key:"timeout,optional"`
}

//  parseURL with parameters
func parseURL(u string) (target, error) {
	rawURL, err := url.Parse(u)
	if err != nil {
		return target{}, errors.Wrap(err, "Malformed URL")
	}

	fmt.Printf("raw url : %s\n", rawURL)

	if rawURL.Scheme != schemeName ||
		len(rawURL.Host) == 0 || len(strings.TrimLeft(rawURL.Path, "/")) == 0 {
		return target{},
			errors.Errorf("Malformed URL('%s'). Must be in the next format: 'polaris://[user:passwd]@host/service?param=value'", u)
	}

	tgt := target{
		Timeout:   time.Duration(500 * time.Millisecond),
		Namespace: "default",
	}
	params := make(map[string]interface{}, len(rawURL.Query()))
	for name, value := range rawURL.Query() {
		params[name] = value[0]
	}

	err = mapping.UnmarshalKey(params, &tgt)
	if err != nil {
		return target{}, errors.Wrap(err, "Malformed URL parameters")
	}

	if tgt.Namespace == "" {
		tgt.Namespace = "default"
	}

	tgt.Addr = rawURL.Host
	tgt.Service = strings.TrimLeft(rawURL.Path, "/")

	fmt.Printf("tgt : %#v\n", tgt)
	return tgt, nil
}
