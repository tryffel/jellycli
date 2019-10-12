/*
 * Copyright 2019 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package components

import "testing"

func TestCoordinates_Limit(t *testing.T) {
	// Test limiting

	c := Rectangle{}
	c.X0 = 10
	c.Y0 = 10

	c.X1 = 20
	c.Y1 = 20

	c.Limit(5, 5)

	x, y := c.Size()
	if x != 5 {
		t.Errorf("Expected x=5, got %d", x)
	}

	if y != 5 {
		t.Errorf("Expected y=5, got %d", y)
	}

	// Test no limit
	c.X0 = 10
	c.Y0 = 10

	c.X1 = 40
	c.Y1 = 40

	c.Limit(50, 50)

	x, y = c.Size()
	if x != 30 {
		t.Errorf("Expected x=30, got %d", x)
	}

	if y != 30 {
		t.Errorf("Expected y=30, got %d", y)
	}

}

func TestRectangle_Limit(t *testing.T) {
	type args struct {
		x int
		y int
	}
	tests := []struct {
		name     string
		rect     Rectangle
		args     args
		wantSize args
	}{
		{
			name: "no-limiting",
			rect: Rectangle{
				X0: 0,
				Y0: 0,
				X1: 10,
				Y1: 10,
			},
			args: args{
				x: 20,
				y: 20,
			},
			wantSize: args{
				x: 10,
				y: 10,
			},
		},
		{
			name: "limit",
			rect: Rectangle{
				X0: 10,
				Y0: 10,
				X1: 30,
				Y1: 30,
			},
			args: args{
				x: 15,
				y: 15,
			},
			wantSize: args{
				x: 15,
				y: 15,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.rect.Limit(tt.args.x, tt.args.y)
			x, y := tt.rect.Size()
			if x != tt.wantSize.x || y != tt.wantSize.y {
				t.Errorf("%s, want (%d, %d), got: (%d, %d)", tt.name, tt.wantSize.x, tt.wantSize.y, x, y)
			}
		})
	}
}

func TestPoint_List(t *testing.T) {
	tests := []struct {
		name   string
		fields Point
		wantX  int
		wantY  int
	}{
		{
			fields: Point{
				X: 10,
				Y: 10,
			},
			wantX: 10,
			wantY: 10,
		},
		{
			fields: Point{
				X: 0,
				Y: 0,
			},
			wantX: 0,
			wantY: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Point{
				X: tt.fields.X,
				Y: tt.fields.Y,
			}
			got, got1 := v.List()
			if got != tt.wantX {
				t.Errorf("List() got = %v, want %v", got, tt.wantX)
			}
			if got1 != tt.wantY {
				t.Errorf("List() got1 = %v, want %v", got1, tt.wantY)
			}
		})
	}
}

func TestRectangle_Sanitize(t *testing.T) {
	tests := []struct {
		name string
		rect Rectangle
		want Rectangle
	}{
		{
			rect: Rectangle{
				X0: -10,
				Y0: -20,
				X1: -30,
				Y1: -40,
			},
			want: Rectangle{
				X0: 0,
				Y0: 0,
				X1: 0,
				Y1: 0,
			},
		},
		{
			rect: Rectangle{
				X0: -10,
				Y0: -20,
				X1: -30,
				Y1: -40,
			},
			want: Rectangle{
				X0: 0,
				Y0: 0,
				X1: 0,
				Y1: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.rect.Sanitize()
			if tt.rect.X0 != tt.want.X0 ||
				tt.rect.X1 != tt.want.X1 ||
				tt.rect.Y0 != tt.want.Y0 ||
				tt.rect.Y1 != tt.want.Y1 {
				t.Errorf("Want (%d,%d)(%d,%d), got: (%d,%d)(%d,%d)",
					tt.want.X0, tt.want.Y0, tt.want.X1, tt.want.Y1,
					tt.rect.X0, tt.rect.Y0, tt.rect.X1, tt.rect.Y1)
			}
		})
	}
}

func TestRectangle_Set(t *testing.T) {
	type fields struct {
		X0 int
		Y0 int
		X1 int
		Y1 int
	}
	type args struct {
		val int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Rectangle{
				X0: tt.fields.X0,
				Y0: tt.fields.Y0,
				X1: tt.fields.X1,
				Y1: tt.fields.Y1,
			}
			_ = c
		})
	}
}

func TestRectangle_Size(t *testing.T) {
	type fields struct {
		X0 int
		Y0 int
		X1 int
		Y1 int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
		want1  int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Rectangle{
				X0: tt.fields.X0,
				Y0: tt.fields.Y0,
				X1: tt.fields.X1,
				Y1: tt.fields.Y1,
			}
			got, got1 := c.Size()
			if got != tt.want {
				t.Errorf("Size() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Size() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
