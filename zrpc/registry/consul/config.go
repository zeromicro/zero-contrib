package consul

import "errors"

const (
	allEths    = "0.0.0.0"
	envPodIP   = "POD_IP"
	consulTags = "consul_tags"
)

// Conf is the config item with the given key on etcd.
type Conf struct {
	Host string
	Key  string
	Tag  []string          `json:",optional"`
	Meta map[string]string `json:",optional"`
}

// Validate validates c.
func (c Conf) Validate() error {
	if len(c.Host) == 0 {
		return errors.New("empty consul hosts")
	} else if len(c.Key) == 0 {
		return errors.New("empty consul key")
	} else {
		return nil
	}
}
