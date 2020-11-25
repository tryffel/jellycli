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
