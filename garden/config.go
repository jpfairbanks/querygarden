package garden

import (
	"errors"
	"os"

	"github.com/jpfairbanks/querygarden/log"
	"github.com/spf13/viper"
)

// Config sets up the viper configuration and reads it into the viper singleton.
func Config(appname string) error {
	log.Debug("loading config for app")
	viper.SetConfigName(appname + "_config")
	viper.AddConfigPath("/etc/" + appname + "/") // path to look for the config file in
	viper.AddConfigPath("$HOME/." + appname)     // call multiple times to add many search paths
	viper.AddConfigPath(".")                     // optionally look for config in the working directory
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
