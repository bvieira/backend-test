package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/caarlos0/env"
)

// Version app
const (
	Version = "1.0.0"
)

type config struct {
	APP  string `env:"JOBS_APP_NAME" envDefault:"c-jobs"`
	Port int    `env:"JOBS_PORT" envDefault:"8080"`

	ElasticSearchServer             string `env:"JOBS_ELASTICSEARCH_SERVER" envDefault:"http://localhost:9200"`
	ElasticSearchMaxRetry           int    `env:"JOBS_ELASTICSEARCH_MAX_RETRY" envDefault:"3"`
	ElasticSearchSniff              bool   `env:"JOBS_ELASTICSEARCH_SNIFF" envDefault:"false"`
	ElasticSearchReconnectRetryTime int    `env:"JOBS_ELASTICSEARCH_RECONNECT_RETRY_TIME_SECONDS" envDefault:"5"`
	ElasticSearchIndexMappingPath   string `env:"JOBS_ELASTICSEARCH_INDEX_MAPPING_PATH" envDefault:"cfg/jobs-mapping.json"`
}

var mutex sync.RWMutex
var cfg *config

//Get get all configuration.
func Get() config {
	mutex.RLock()
	defer mutex.RUnlock()
	return *cfg
}

func init() {
	Load()
}

//Load it loads the configuration from property
func Load() {
	mutex.Lock()
	defer mutex.Unlock()
	cfg = &config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	//updates environment variables with default if not set
	os.Setenv("JOBS_APP_VERSION", Version)
	v := reflect.ValueOf(cfg).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if e := field.Tag.Get("env"); e != "" {
			os.Setenv(e, fmt.Sprintf("%v", v.Field(i)))
		}
	}
}

//List configs
func List() []Env {
	var result []Env
	v := reflect.ValueOf(&config{}).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if e := field.Tag.Get("env"); e != "" {
			result = append(result, Env{Name: e, Type: field.Type.String(), DefaultVal: field.Tag.Get("envDefault")})
		}
	}
	return result
}

type Env struct {
	Name       string
	Type       string
	DefaultVal string
}
