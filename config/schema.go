package config

type ServiceConfig struct {
	TTL      int64  `yaml:"ttl"`
	Interval int64  `yaml:"interval"`
	Address  string `yaml:"address"`
}

type LoggerConfig struct {
	Level string `yaml:"level"`
	Dir string `yaml:"dir"`
}

type DBConfig struct {
	Type     string	`yaml:"type"`
	User     string	`yaml:"user"`
	Password string	`yaml:"password"`
	IP      string	`yaml:"ip"`
	Port     string	`yaml:"port"`
	Name     string	`yaml:"name"`
}

type GraphConfig struct {
	Password string	`yaml:"password"`
	IP      string	`yaml:"ip"`
	Port     string	`yaml:"port"`
	Name     string	`yaml:"name"`
	User     string	`yaml:"user"`
}

type BasicConfig struct {
	SynonymMax    int32 `yaml:"synonyms"`
	TagMax    	  int32 `yaml:"tags"`
}

type SchemaConfig struct {
	Service  ServiceConfig `yaml:"service"`
	Logger   LoggerConfig  `yaml:"logger"`
	Database DBConfig      `yaml:"database"`
	Graph    GraphConfig   `yaml:"graph"`
	//Basic   BasicConfig 	`yaml:"basic"`
}
