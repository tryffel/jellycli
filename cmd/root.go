/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cmd

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"path"
	"strings"
	"sync"
	"tryffel.net/go/jellycli/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Long: `Jellycli is a terminal music player for
Jellyfin and Subsonic-compatible servers.

`,

	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		runApplication()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.Flags().BoolVar(&disableGui, "no-gui", false, "disable gui")
}

func initConfig() {
	// default config dir is ~/.config/jellycli
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			logrus.Errorf("cannot determine config directory: %v", err)
			configDir = ""
		} else {
			configDir = path.Join(configDir, "jellycli")
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigFile(path.Join(configDir, "jellycli.yaml"))
	}

	// env variables
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvPrefix("jellycli")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = config.NewConfigFile(cfgFile)
			if err != nil {
				logrus.Fatalf("create config file: %v", err)
			}
		} else {
			logrus.Fatalf("read config file: %v", err)
		}
	}

	err := config.ConfigFromViper()
	if err != nil {
		logrus.Fatalf("read config file: %v", err)
	}
}

func initLogging() (*os.File, error) {
	level, err := logrus.ParseLevel(config.AppConfig.Player.LogLevel)
	if err != nil {
		logrus.Errorf("parse log level: %v", err)
		return nil, nil
	}

	logrus.SetLevel(level)
	format := &prefixed.TextFormatter{
		ForceColors:      false,
		DisableColors:    true,
		ForceFormatting:  true,
		DisableTimestamp: false,
		DisableUppercase: false,
		FullTimestamp:    true,
		TimestampFormat:  "",
		DisableSorting:   false,
		QuoteEmptyFields: false,
		QuoteCharacter:   "'",
		SpacePadding:     0,
		Once:             sync.Once{},
	}
	logrus.SetFormatter(format)
	dir := os.TempDir()
	file := path.Join(dir, fmt.Sprintf("%s.log", strings.ToLower(config.AppName)))
	fd, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0760))
	if err != nil {
		return nil, fmt.Errorf("open log file: %v", err)
	}
	config.LogFile = file
	logrus.SetOutput(fd)
	return fd, nil
}
