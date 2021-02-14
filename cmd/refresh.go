/*
 * Jellycli is a terminal music player for Jellyfin.
 * Copyright (C) 2020 Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"tryffel.net/go/jellycli/config"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Pull latest data from remote server and store to local cache",
	Run: func(cmd *cobra.Command, args []string) {
		disableGui = true
		initConfig()
		logFile, err := initLogging()
		if err != nil {
			logrus.Fatalf("init logging: %v", err)
		}

		a := &app{}
		a.logfile = logFile
		// write to both log file and stdout for startup, in case there are any errors that prevent gui
		writer := io.MultiWriter(a.logfile, os.Stdout)
		logrus.SetOutput(writer)

		logrus.Infof("############# %s v%s ############", config.AppName, config.Version)

		if !config.AppConfig.Player.EnableLocalCache {
			logrus.Fatalf("Local cache is disabled")
		}

		err = a.initServerConnection()
		if err != nil {
			logrus.Fatalf("connect to server: %v", err)
		}

		err = a.initApp()
		if err != nil {
			logrus.Fatalf("init application: %v", err)
		}

		ok := false

		quit := func() {
			a.player.Stop()
			a.server.Stop()
			a.logfile.Close()
			if !ok {
				os.Exit(1)
			}
		}
		defer quit()

		err = a.player.UpdateLocalArtists(0)
		if err != nil {
			logrus.Error(err)
			return
		}

		err = a.player.UpdateLocalAlbums(0)
		if err != nil {
			logrus.Error(err)
			return
		}

		err = a.player.UpdateLocalSongs(0)
		if err != nil {
			logrus.Error(err)
			return
		}

		err = a.player.UpdatePlaylists()
		if err != nil {
			logrus.Error(err)
			return
		}
		ok = true
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
}
