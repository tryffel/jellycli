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

package widgets

import (
	"gitlab.com/tslocum/cview"
	"reflect"
	"testing"
	"tryffel.net/go/twidgets"
)

// dummy struct to implement twidgets.ListItme and store index we can compare.
type testListItem struct {
	*cview.Box
	index int
}

func newTestListItem(index int) *testListItem {
	return &testListItem{
		Box:   nil,
		index: index,
	}
}

var texts = []string{
	"Lorem ipsum dolor sit amet",
	"consectetur adipiscing elit",
	"sed do eiusmod tempor incididunt",
	"ut labore et dolore magna aliqua.",
	"Ut enim",
	"ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip",
	"ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate ",
	"velit esse cillum dolore eu fugiat nulla pariatur.",
	"Excepteur sint occaecat cupidatat non proident, sunt in ",
	"culpa qui officia deserunt mollit anim id est laborum.",
}

func (t testListItem) SetSelected(selected twidgets.Selection) {}

func Test_itemList_reduce(t *testing.T) {
	list := newItemList(nil)
	list.itemsTexts = texts

	list.items = []twidgets.ListItem{
		newTestListItem(0),
		newTestListItem(1),
		newTestListItem(2),
		newTestListItem(3),
		newTestListItem(4),
		newTestListItem(5),
		newTestListItem(6),
		newTestListItem(7),
		newTestListItem(8),
		newTestListItem(9),
	}
	tests := []struct {
		name        string
		input       string
		wantIndices []int
	}{
		{
			name:        "simple match",
			input:       "ipsum",
			wantIndices: []int{0},
		},
		{
			name:        "empty query",
			input:       "",
			wantIndices: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:        "invalid query",
			input:       " ., .",
			wantIndices: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name:        "uppercased query",
			input:       "AD",
			wantIndices: []int{1, 5},
		},
		{
			name:        "multi-case query",
			input:       "AD ven",
			wantIndices: []int{5},
		},
		{
			name:        "multi-case no match",
			input:       "AD moc",
			wantIndices: []int{},
		},
		{
			name:        "multi-case with stripped characters",
			input:       "AD + & ven",
			wantIndices: []int{5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			list.reduce(tt.input)

			if !reflect.DeepEqual(list.reduceIndices, tt.wantIndices) {
				t.Errorf("reduced item indices do not match")
			}
		})
	}
}

// some benchmarks for filtering results.
// Intel i7-2600K @ 3.40GHz
// Name											   total rounds        time
// Benchmark_itemList_reduce100items-8          	  167535	      8003 ns/op
// Benchmark_itemList_reduce1000items-8         	   23006	     48804 ns/op
// Benchmark_itemList_reduce10000items-8        	    2289	    493618 ns/op
// Benchmark_itemList_reduce10000largeItems-8   	     468	   2480820 ns/op
//
// So, with reasonable page size (<1000 items) and normal items (not articles as song names)
// filtering should be fast enough.

func Benchmark_itemList_reduce100items(b *testing.B) {
	b.StopTimer()
	itemCount := 100

	list := newItemList(nil)
	testItem := newTestListItem(0)
	items := make([]string, len(texts)*10)
	for i := 0; i < itemCount; i++ {
		items[i] = texts[i%10]
		list.items = append(list.items, testItem)
	}

	list.itemsTexts = items
	input := "con"

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		list.reduce(input)
	}
}

func Benchmark_itemList_reduce1000items(b *testing.B) {
	b.StopTimer()
	itemCount := 1000

	list := newItemList(nil)
	testItem := newTestListItem(0)
	items := make([]string, len(texts)*100)
	for i := 0; i < itemCount; i++ {
		items[i] = texts[i%10]
		list.items = append(list.items, testItem)
	}

	list.itemsTexts = items
	input := "con"

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		list.reduce(input)
	}
}

func Benchmark_itemList_reduce10000items(b *testing.B) {
	b.StopTimer()
	itemCount := 10000

	list := newItemList(nil)
	testItem := newTestListItem(0)
	items := make([]string, len(texts)*itemCount/len(texts))
	for i := 0; i < itemCount; i++ {
		items[i] = texts[i%10]
		list.items = append(list.items, testItem)
	}

	list.itemsTexts = items
	input := "con"

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		list.reduce(input)
	}
}

func Benchmark_itemList_reduce10000largeItems(b *testing.B) {
	// 8 times longer item names for each item.
	b.StopTimer()
	itemCount := 10000

	list := newItemList(nil)
	testItem := newTestListItem(0)
	items := make([]string, len(texts)*itemCount/len(texts))
	for i := 0; i < itemCount; i++ {
		items[i] = texts[i%10] + texts[i%10] + texts[i%10] + texts[i%10] + texts[i%10] + texts[i%10] + texts[i%10] + texts[i%10]
		list.items = append(list.items, testItem)
	}

	list.itemsTexts = items
	input := "con"

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		list.reduce(input)
	}
}
