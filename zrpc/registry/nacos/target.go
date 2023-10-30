package nacos

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/mapping"
)

type target struct {
	Addr        string        `key:",optional"`
	User        string        `key:",optional"`
	Password    string        `key:",optional"`
	Service     string        `key:",optional"`
	GroupName   string        `key:",optional"`
	Clusters    []string      `key:",optional"`
	NamespaceID string        `key:"namespaceid,optional"`
	Timeout     time.Duration `key:"timeout,optional"`
	AppName     string        `key:"appName,optional"`
	LogLevel    string        `key:",optional"`
	LogDir      string        `key:",optional"`
	CacheDir    string        `key:",optional"`
	NotLoadCacheAtStart  bool `key:"notLoadCacheAtStart,optional"`
	UpdateCacheWhenEmpty bool `key:"updateCacheWhenEmpty,optional"`
}

// parseURL with parameters
func parseURL(rawURL url.URL) (target, error) {
	if rawURL.Scheme != schemeName ||
		len(rawURL.Host) == 0 || len(strings.TrimLeft(rawURL.Path, "/")) == 0 {
		return target{},
			errors.Errorf("Malformed URL('%s'). Must be in the next format: 'nacos://[user:passwd]@host/service?param=value'", rawURL.String())
	}

	var tgt target
	params := make(map[string]interface{}, len(rawURL.Query()))
	for name, value := range rawURL.Query() {
		params[name] = value[0]
	}

	err := mapping.UnmarshalKey(params, &tgt)
	if err != nil {
		return target{}, errors.Wrap(err, "Malformed URL parameters")
	}

	if tgt.NamespaceID == "" {
		tgt.NamespaceID = "public"
	}

	tgt.LogLevel = os.Getenv("NACOS_LOG_LEVEL")
	tgt.LogDir = os.Getenv("NACOS_LOG_DIR")
	tgt.CacheDir = os.Getenv("NACOS_CACHE_DIR")

	tgt.User = rawURL.User.Username()
	tgt.Password, _ = rawURL.User.Password()
	tgt.Addr = rawURL.Host
	tgt.Service = strings.TrimLeft(rawURL.Path, "/")

	if logLevel, exists := os.LookupEnv("NACOS_LOG_LEVEL"); exists {
		tgt.LogLevel = logLevel
	}

	if logDir, exists := os.LookupEnv("NACOS_LOG_DIR"); exists {
		tgt.LogDir = logDir
	}

	if notLoadCacheAtStart, exists := os.LookupEnv("NACOS_NOT_LOAD_CACHE_AT_START"); exists {
		tgt.NotLoadCacheAtStart = notLoadCacheAtStart == "true"
	}

	if updateCacheWhenEmpty, exists := os.LookupEnv("NACOS_UPDATE_CACHE_WHEN_EMPTY"); exists {
		tgt.UpdateCacheWhenEmpty = updateCacheWhenEmpty == "true"
	}

	return tgt, nil
}
