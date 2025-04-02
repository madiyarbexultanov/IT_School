package config

var Config *MapConfig

type MapConfig struct {
	AppHost            string `mapstructure:"APP_HOST"`
	DbConnectionString string `mapstructure:"DB_CONNECTION_STRING"`
}
