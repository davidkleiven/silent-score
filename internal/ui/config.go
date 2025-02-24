package ui

type Config struct {
	DbName  string
	LogFile string
}

func defaultConfig() *Config {
	return &Config{
		DbName:  "silent-score.db",
		LogFile: "silent-score.log",
	}
}

type EditConfigFunc func(c *Config)

func WithDb(dbName string) EditConfigFunc {
	return func(c *Config) {
		c.DbName = dbName
	}
}

func WithLogFile(name string) EditConfigFunc {
	return func(c *Config) {
		c.LogFile = name
	}
}

func NewConfig(edits ...EditConfigFunc) *Config {
	c := defaultConfig()
	for _, editFunc := range edits {
		editFunc(c)
	}
	return c
}
