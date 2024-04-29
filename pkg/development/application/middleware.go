package application

type Middleware struct {
	Postgres   *PostgresConfig   `yaml:"postgres,omitempty" json:"postgres,omitempty"`
	Redis      *RedisConfig      `yaml:"redis,omitempty" json:"redis,omitempty"`
	MongoDB    *MongodbConfig    `yaml:"mongodb,omitempty" json:"mongodb,omitempty"`
	ZincSearch *ZincSearchConfig `yaml:"zincSearch,omitempty" json:"zincSearch,omitempty"`
}

type Database struct {
	Name        string `json:"name"`
	Distributed bool   `json:"distributed,omitempty"`
}

type PostgresConfig struct {
	Username  string     `yaml:"username" json:"username"`
	Password  string     `yaml:"password,omitempty" json:"password,omitempty"`
	Databases []Database `yaml:"databases" json:"databases"`
}

type RedisConfig struct {
	Password  string `yaml:"password,omitempty" json:"password,omitempty"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

type MongodbConfig struct {
	Username  string     `yaml:"username" json:"username"`
	Password  string     `yaml:"password,omitempty" json:"password,omitempty"`
	Databases []Database `yaml:"databases" json:"databases"`
}

type ZincSearchConfig struct {
	Username string  `yaml:"username" json:"username"`
	Password string  `yaml:"password,omitempty" json:"password,omitempty"`
	Indexes  []Index `yaml:"indexes" json:"indexes"`
}

type Index struct {
	Name string `yaml:"name" json:"name"`
}
