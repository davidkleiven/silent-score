package ui

import "testing"

func configIsEqual(c1, c2 *Config) bool {
	return c1.DbName == c2.DbName
}

func TestInitNoEdits(t *testing.T) {
	config := NewConfig()
	defaultCfg := defaultConfig()
	if !configIsEqual(config, defaultCfg) {
		t.Errorf("Wanted %+v got %+v\n", defaultCfg, config)
	}

}

func TestSetDbName(t *testing.T) {
	name := "my-custom-database.db"
	config := NewConfig(WithDb(name))
	if config.DbName != name {
		t.Errorf("Wanted %s got %s\n", name, config.DbName)
	}
}

func TestSetWithLogFile(t *testing.T) {
	config := NewConfig(WithLogFile("logfile.log"))
	if config.LogFile != "logfile.log" {
		t.Errorf("Wanted logfile.log got %s", config.LogFile)
	}
}
