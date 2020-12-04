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

package jellyfin

import (
	"reflect"
	"testing"
	"tryffel.net/go/jellycli/interfaces"
)

func Test_params_setPaging(t *testing.T) {
	type args struct {
		paging interfaces.Paging
	}
	tests := []struct {
		name string
		p    params
		args args
		want map[string]string
	}{
		{
			p: params{},
			args: args{paging: interfaces.Paging{
				TotalItems:  3500,
				TotalPages:  035,
				CurrentPage: 4,
				PageSize:    100,
			}},
			want: map[string]string{"Limit": "100", "StartIndex": "400"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.setPaging(tt.args.paging)

			reflect.DeepEqual(&tt, tt.want)
		})
	}
}
