package config

import "time"

var Config *MapConfig

type MapConfig struct {
	DbConnectionString string  		 `mapstructure:"DATABASE_URL"`
	JwtSecretKey       string  		 `mapstructure:"JWT_SECRET_KEY"`
	JwtExpiresIn       time.Duration 	 `mapstructure:"JWT_EXPIRE_DURATION"`
	SMTPEmail   	   string 		 `mapstructure:"SMTP_EMAIL"`
    	SMTPPassword 	   string 		 `mapstructure:"SMTP_PASSWORD"`
    	SMTPHost     	   string 		 `mapstructure:"SMTP_HOST"`
    	SMTPPort     	   string 		 `mapstructure:"SMTP_PORT"`
    	Initial_Password   string 		 `mapstructure:"INITIAL_PASSWORD"`
   	Admin_Name   	   string 		 `mapstructure:"ADMIN_NAME"`
    	Admin_Mail   	   string 		 `mapstructure:"ADMIN_MAIL"`
    	Admin_Phone   	   string 		 `mapstructure:"ADMIN_PHONE"`
}
