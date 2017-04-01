package featex

import (
	"errors"
	"os"

	"github.com/jpfairbanks/featex/log"
	"github.com/spf13/viper"
)

// Config sets up the viper configuration and reads it into the viper singleton.
func Config() error {
	viper.SetConfigName("featex_config")
	viper.AddConfigPath("/etc/featex/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.featex") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	defglobal := map[string]interface{}{
		"schema":  "schema",
		"version": "1.0",
	}
	viper.SetDefault("global", defglobal)
	err := viper.ReadInConfig()
	if err != nil {
		log.Error(err)
		return errors.New("Cannot load configuration check for file featex.config in ./, /etc/featex/, or $HOME/.featex/")
	}
	return nil
}

// DBString gets the database configuration from the environment.
func DBString() string {
	dbstring, ok := os.LookupEnv("DBSTRING")
	if !ok {
		dbstring = "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
	}
	return dbstring
}
