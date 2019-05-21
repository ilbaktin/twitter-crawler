package conf

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

var lock sync.Mutex
var configFile string

type PostgresAccessConfig struct {
	Host		string		`yaml:"host"`
	Port		*uint16		`yaml:"port,omitempty"`
	Dbname		string		`yaml:"dbname"`
	User		string		`yaml:"user"`
	Password	*string		`yaml:"password,omitempty"`
}

type MasterConfig struct {
	NumOfWorkers		int							`yaml:"num_of_workers"`
	PostgresAccess		PostgresAccessConfig		`yaml:"pg_access"`
	Cookies				map[string]string			`yaml:"cookies"`
	Headers				map[string]string			`yaml:"headers"`
}

func Init(configFilePath string) {
	configFile = configFilePath
}

func LoadConfig() (*MasterConfig, error) {
	lock.Lock()
	defer lock.Unlock()
	yamlBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := &MasterConfig{}
	err = yaml.Unmarshal(yamlBytes, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
