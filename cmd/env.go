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

import "github.com/spf13/cobra"

var envCmd = &cobra.Command{
	Use:   "list-env",
	Short: "List env variables",
	Long: `Any configuration variable can be set with environment variables. In addition,
it is also possible to define passwords for servers. This way it would be possible to use
Jellycli without persisting config file (with e.g. Docker). Jellycli will still create config file, nevertheless.

JELLYCLI_JELLYFIN_URL
JELLYCLI_JELLYFIN_PASSWORD
JELLYCLI_JELLYFIN_TOKEN
JELLYCLI_JELLYFIN_USERID
JELLYCLI_JELLYFIN_DEVICE_ID
JELLYCLI_JELLYFIN_SERVER_ID
JELLYCLI_JELLYFIN_MUSIC_VIEW

JELLYCLI_SUBSONIC_URL
JELLYCLI_SUBSONIC_USERNAME
JELLYCLI_SUBSONIC_PASSWORD
JELLYCLI_SUBSONIC_SALT
JELLYCLI_SUBSONIC_TOKEN

JELLYCLI_PLAYER_SERVER
JELLYCLI_PLAYER_PAGESIZE
JELLYCLI_PLAYER_LOGFILE
JELLYCLI_PLAYER_LOGLEVEL
JELLYCLI_PLAYER_DEBUG_MODE
JELLYCLI_PLAYER_LIMIT_RECENTLY_PLAYED
JELLYCLI_PLAYER_MOUSE_ENABLED
JELLYCLI_PLAYER_AUDIO_BUFFERING_MS
JELLYCLI_PLAYER_DOUBLE_CLICK_MS
JELLYCLI_PLAYER_HTTP_BUFFERING_S
JELLYCLI_PLAYER_HTTP_BUFFERING_LIMIT_MEM
JELLYCLI_PLAYER_ENABLE_REMOTE_CONTROL
JELLYCLI_PLAYER_SEARCH_RESULTS_LIMIT

`,
}

func init() {
	rootCmd.AddCommand(envCmd)

}
