package settings

/*
	@Author: Sinh
	@Date: 2025/6/1
	@Description: This file create a configuration struct for the application.
	@Note: The configuration struct is used to store the configuration of the application.
	All the configuration is loaded from a YAML file using Viper.
	And it is stored in the global variable `Config` in the `global` package.
*/

type Config struct {
	Server   Server   `mapstructure:"server" json:"server" yaml:"server"`
	Redis    Redis    `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log      Log      `mapstructure:"log" json:"log" yaml:"log"`
	RabbitMQ RabbitMQ `mapstructure:"rabbitmq" json:"rabbitmq" yaml:"rabbitmq"`
	Mongo    Mongo    `mapstructure:"mongo" json:"mongo" yaml:"mongo"`
}

type Redis struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	DB       int    `mapstructure:"db" json:"db" yaml:"db"`
	PoolSize int    `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`
	MinIdle  int    `mapstructure:"min_idle" json:"min_idle" yaml:"min_idle"`
}

type Log struct {
	Level string `mapstructure:"level" json:"level" yaml:"level"`
}

type RabbitMQ struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	User     string `mapstructure:"user" json:"user" yaml:"user"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
}

type Server struct {
	Host                 string `mapstructure:"host" json:"host" yaml:"host"`
	PersonalSchedulePort int    `mapstructure:"personal_schedule_port" json:"personal_schedule_port" yaml:"personal_schedule_port"`
	MaxRecvMsgSize       int    `mapstructure:"max_recv_msg_size" json:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	MaxSendMsgSize       int    `mapstructure:"max_send_msg_size" json:"max_send_msg_size" yaml:"max_send_msg_size"`
	KeepaliveTime        int    `mapstructure:"keepalive_time" json:"keepalive_time" yaml:"keepalive_time"`          // in seconds
	KeepaliveTimeout     int    `mapstructure:"keepalive_timeout" json:"keepalive_timeout" yaml:"keepalive_timeout"` // in seconds
}

type Mongo struct {
	URI      string `mapstructure:"uri" json:"uri" yaml:"uri"`
	Database string `mapstructure:"database" json:"database" yaml:"database"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
}
