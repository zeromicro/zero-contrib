package polaris

const (
	allEths  = "0.0.0.0"
	envPodIP = "POD_IP"
)

// options
type Options struct {
	ListenOn            string
	Namespace           string
	ServiceToken        string
	ServiceName         string
	Weight              float64
	Protocol            string
	Version             string
	HeartbeatInervalSec int
	Metadata            map[string]string
}

type Option func(*Options)

func NewPolarisConfig(listenOn string, opts ...Option) *Options {
	options := &Options{
		ListenOn:            listenOn,
		Namespace:           "default",
		Protocol:            "zrpc",
		Version:             "1.0.0",
		HeartbeatInervalSec: 5,
		Metadata:            make(map[string]string),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func WithHeartbeatInervalSec(heartbeatInervalSec int) Option {
	return func(o *Options) {
		o.HeartbeatInervalSec = heartbeatInervalSec
	}
}

func WithWeight(weight float64) Option {
	return func(o *Options) {
		o.Weight = weight
	}
}

func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

func WithServiceName(serviceName string) Option {
	return func(o *Options) {
		o.ServiceName = serviceName
	}
}

func WithVersion(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

func WithProtocol(protocol string) Option {
	return func(o *Options) {
		o.Protocol = protocol
	}
}

func WithMetadata(metadata map[string]string) Option {
	return func(o *Options) {
		o.Metadata = metadata
	}
}
