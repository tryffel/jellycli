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
	"fmt"
	"github.com/spf13/cobra"
	"tryffel.net/go/jellycli/config"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Jellycli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", config.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
