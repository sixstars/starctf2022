package ldap

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/BurntSushi/toml"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util/errutil"
)

// Config holds list of connections to LDAP
type Config struct {
	Servers []*ServerConfig `toml:"servers"`
}

// ServerConfig holds connection data to LDAP
type ServerConfig struct {
	Host          string       `toml:"host"`
	Port          int          `toml:"port"`
	UseSSL        bool         `toml:"use_ssl"`
	StartTLS      bool         `toml:"start_tls"`
	SkipVerifySSL bool         `toml:"ssl_skip_verify"`
	RootCACert    string       `toml:"root_ca_cert"`
	ClientCert    string       `toml:"client_cert"`
	ClientKey     string       `toml:"client_key"`
	BindDN        string       `toml:"bind_dn"`
	BindPassword  string       `toml:"bind_password"`
	Attr          AttributeMap `toml:"attributes"`

	SearchFilter  string   `toml:"search_filter"`
	SearchBaseDNs []string `toml:"search_base_dns"`

	GroupSearchFilter              string   `toml:"group_search_filter"`
	GroupSearchFilterUserAttribute string   `toml:"group_search_filter_user_attribute"`
	GroupSearchBaseDNs             []string `toml:"group_search_base_dns"`

	Groups []*GroupToOrgRole `toml:"group_mappings"`
}

// AttributeMap is a struct representation for LDAP "attributes" setting
type AttributeMap struct {
	Username string `toml:"username"`
	Name     string `toml:"name"`
	Surname  string `toml:"surname"`
	Email    string `toml:"email"`
	MemberOf string `toml:"member_of"`
}

// GroupToOrgRole is a struct representation of LDAP
// config "group_mappings" setting
type GroupToOrgRole struct {
	GroupDN string `toml:"group_dn"`
	OrgId   int64  `toml:"org_id"`

	// This pointer specifies if setting was set (for backwards compatibility)
	IsGrafanaAdmin *bool `toml:"grafana_admin"`

	OrgRole models.RoleType `toml:"org_role"`
}

// logger for all LDAP stuff
var logger = log.New("ldap")

// loadingMutex locks the reading of the config so multiple requests for reloading are sequential.
var loadingMutex = &sync.Mutex{}

// IsEnabled checks if ldap is enabled
func IsEnabled() bool {
	return setting.LDAPEnabled
}

// ReloadConfig reads the config from the disk and caches it.
func ReloadConfig() error {
	if !IsEnabled() {
		return nil
	}

	loadingMutex.Lock()
	defer loadingMutex.Unlock()

	var err error
	config, err = readConfig(setting.LDAPConfigFile)
	return err
}

// We need to define in this space so `GetConfig` fn
// could be defined as singleton
var config *Config

// GetConfig returns the LDAP config if LDAP is enabled otherwise it returns nil. It returns either cached value of
// the config or it reads it and caches it first.
func GetConfig(cfg *setting.Cfg) (*Config, error) {
	if cfg != nil {
		if !cfg.LDAPEnabled {
			return nil, nil
		}
	} else if !IsEnabled() {
		return nil, nil
	}

	// Make it a singleton
	if config != nil {
		return config, nil
	}

	loadingMutex.Lock()
	defer loadingMutex.Unlock()

	return readConfig(setting.LDAPConfigFile)
}

func readConfig(configFile string) (*Config, error) {
	result := &Config{}

	logger.Info("LDAP enabled, reading config file", "file", configFile)

	// nolint:gosec
	// We can ignore the gosec G304 warning on this one because `filename` comes from grafana configuration file
	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, errutil.Wrap("Failed to load LDAP config file", err)
	}

	// interpolate full toml string (it can contain ENV variables)
	stringContent, err := setting.ExpandVar(string(fileBytes))
	if err != nil {
		return nil, errutil.Wrap("Failed to expand variables", err)
	}

	_, err = toml.Decode(stringContent, result)
	if err != nil {
		return nil, errutil.Wrap("Failed to load LDAP config file", err)
	}

	if len(result.Servers) == 0 {
		return nil, fmt.Errorf("LDAP enabled but no LDAP servers defined in config file")
	}

	// set default org id
	for _, server := range result.Servers {
		err = assertNotEmptyCfg(server.SearchFilter, "search_filter")
		if err != nil {
			return nil, errutil.Wrap("Failed to validate SearchFilter section", err)
		}
		err = assertNotEmptyCfg(server.SearchBaseDNs, "search_base_dns")
		if err != nil {
			return nil, errutil.Wrap("Failed to validate SearchBaseDNs section", err)
		}

		for _, groupMap := range server.Groups {
			if groupMap.OrgId == 0 {
				groupMap.OrgId = 1
			}
		}
	}

	return result, nil
}

func assertNotEmptyCfg(val interface{}, propName string) error {
	switch v := val.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("LDAP config file is missing option: %q", propName)
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("LDAP config file is missing option: %q", propName)
		}
	default:
		fmt.Println("unknown")
	}
	return nil
}
