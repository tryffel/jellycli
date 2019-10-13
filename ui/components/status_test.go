package components

import "testing"

func Test_secToString(t *testing.T) {
	type args struct {
		sec int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{sec: 5},
			want: "0:05",
		},
		{
			args: args{sec: 45},
			want: "0:45",
		},
		{
			args: args{sec: 90},
			want: "1:30",
		},
		{
			args: args{sec: 65},
			want: "1:05",
		},
		{
			args: args{sec: 1515},
			want: "25:15",
		},
		{
			args: args{sec: 3820},
			want: "1:03:40",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := secToString(tt.args.sec); got != tt.want {
				t.Errorf("secToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
