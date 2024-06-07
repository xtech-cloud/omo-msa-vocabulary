package config

type ServiceConfig struct {
	TTL      int64  `json:"ttl"`
	Interval int64  `json:"interval"`
	Address  string `json:"address"`
}

type LoggerConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
	Std   bool   `json:"std"`
}

type DBConfig struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
	Name     string `json:"name"`
}

type GraphConfig struct {
	Password string `json:"password"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
}

type GraphType struct {
	Type uint8  `json:"type"`
	Name string `json:"name"`
}

type BasicConfig struct {
	SynonymMax int32        `json:"synonyms"`
	TagMax     int32        `json:"tags"`
	Kinds      []*GraphType `json:"kinds"`
}

type SchemaConfig struct {
	Service  ServiceConfig `json:"service"`
	Logger   LoggerConfig  `json:"logger"`
	Database DBConfig      `json:"database"`
	Graph    GraphConfig   `json:"graph"`
	Basic    BasicConfig   `json:"basic"`
}

func (mine *BasicConfig) GetName(tp uint8) string {
	for _, kind := range mine.Kinds {
		if kind.Type == tp {
			return kind.Name
		}
	}
	return ""
}
