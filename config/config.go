package config

import "time"

var Config *MapConfig

type MapConfig struct {
	AppHost            string  		 `mapstructure:"APP_HOST"`
	DbConnectionString string  		 `mapstructure:"DB_CONNECTION_STRING"`
	JwtSecretKey       string  		 `mapstructure:"JWT_SECRET_KEY"`
	JwtExpiresIn       time.Duration `mapstructure:"JWT_EXPIRE_DURATION"`
	SMTPEmail   	   string 		 `mapstructure:"SMTP_EMAIL"`
    SMTPPassword 	   string 		 `mapstructure:"SMTP_PASSWORD"`
    SMTPHost     	   string 		 `mapstructure:"SMTP_HOST"`
    SMTPPort     	   string 		 `mapstructure:"SMTP_PORT"`
}
