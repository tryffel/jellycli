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
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
	"regexp"
	"strconv"
	"strings"
	"tryffel.net/go/jellycli/config"
	"tryffel.net/go/jellycli/interfaces"
	"tryffel.net/go/jellycli/ui/widgets/modal"
)

const (
	sortAsc  = " \U0001F883 "
	sortDesc = "\U0001F881 "
)

type queryFunc = func(opts *interfaces.QueryOpts)
type sortFunc = func(sort interfaces.Sort)
type openFilterFunc = func(m modal.Modal, doneFunc func())

type sort struct {
	*dropDown
	currentIndex int
	mode         interfaces.SortMode

	sortFunc sortFunc
}

func newSort(sortFunc sortFunc, options ...interfaces.SortField) *sort {
	s := &sort{
		dropDown:     newDropDown("Sort"),
		currentIndex: 0,
		mode:         interfaces.SortAsc,
		sortFunc:     sortFunc,
	}

	s.SetTextOptions("", "", sortAsc, "", "")

	for i, v := range options {
		s.AddOption(string(v), func() {
			s.setSorting(i, v)
		})
	}
	return s
}

func (s *sort) setSorting(newIndex int, field interfaces.SortField) {
	if newIndex == s.currentIndex {
		s.toggleMode()
	} else {
		s.SetTextOptions("", "", sortAsc, "", "")
		s.currentIndex = newIndex
	}

	if s.sortFunc != nil {
		sort := interfaces.Sort{Mode: string(s.mode), Field: field}
		s.sortFunc(sort)
	}
}

func (s *sort) toggleMode() {
	if s.mode == interfaces.SortAsc {
		s.mode = interfaces.SortDesc
		s.SetTextOptions("", "", sortDesc, "", "")
	} else {
		s.mode = interfaces.SortAsc
		s.SetTextOptions("", "", sortAsc, "", "")
	}
}

type filter struct {
	*cview.Form
	filterFunc func(interfaces.Filter)

	visible bool
	closeCb func()

	itemPlayed    *cview.Checkbox
	itemNotPlayed *cview.Checkbox

	itemFavorite *cview.Checkbox

	yearRange *cview.InputField
}

func (f *filter) SetDoneFunc(doneFunc func()) {
	f.closeCb = doneFunc
}

func (f *filter) View() cview.Primitive {
	return f
}

func (f *filter) SetVisible(visible bool) {
	f.visible = visible
}

func newFilter(itemType string, filterFunc func(f interfaces.Filter)) *filter {

	f := &filter{
		Form:       cview.NewForm(),
		filterFunc: filterFunc,

		itemPlayed:    cview.NewCheckbox(),
		itemNotPlayed: cview.NewCheckbox(),
		itemFavorite:  cview.NewCheckbox(),
		yearRange:     cview.NewInputField(),
	}

	f.SetTitle(fmt.Sprintf(" Filter %ss ", itemType))

	f.SetBackgroundColor(config.Color.Modal.Background)
	f.SetBorder(true)
	f.AddFormItem(f.itemPlayed)
	f.AddFormItem(f.itemNotPlayed)

	f.itemPlayed.SetDoneFunc(excludeOther(f.itemPlayed, f.itemNotPlayed))
	f.itemNotPlayed.SetDoneFunc(excludeOther(f.itemNotPlayed, f.itemPlayed))

	f.yearRange.SetAcceptanceFunc(validateYearRange)

	f.itemPlayed.SetLabel("Played")
	f.itemNotPlayed.SetLabel("Not played")
	f.itemFavorite.SetLabel("Favorite")
	f.yearRange.SetLabel("Year range")
	f.yearRange.SetPlaceholder("'2020' or '2000-2010'")
	f.yearRange.SetPlaceholderTextColor(config.Color.TextDisabled)
	f.yearRange.SetFieldTextColor(config.Color.Text)

	f.AddFormItem(f.itemFavorite)
	f.AddFormItem(f.yearRange)

	f.AddButton("Filter", f.ok)
	f.AddButton("Cancel", f.closeCb)

	return f
}

func excludeOther(c1, c2 *cview.Checkbox) func(key tcell.Key) {
	// set two checkboxes exclusive
	return func(key tcell.Key) {
		if c1.IsChecked() {
			c2.SetChecked(false)
		}
	}
}

var validYearRangeRe = regexp.MustCompile("^([0-9]{1,4})$")

func validateYearRange(textToCheck string, lastChar rune) bool {
	splitchar := "-"

	if !strings.Contains(textToCheck, splitchar) {
		return validYearRangeRe.MatchString(textToCheck)

	}

	splits := strings.Split(textToCheck, "-")
	if len(splits) == 1 {
		return validYearRangeRe.MatchString(splits[0])
	}

	if len(splits) == 2 {
		if splits[1] == "" {
			return validYearRangeRe.MatchString(splits[0])
		}
		return validYearRangeRe.MatchString(splits[0]) && validYearRangeRe.MatchString(splits[1])
	} else {
		return false
	}
}

func (f *filter) ok() {

	if f.filterFunc == nil {
		f.closeCb()
	}

	filt := interfaces.Filter{
		FilterPlayed: "",
		Favorite:     f.itemFavorite.IsChecked(),
		Genres:       nil,
		YearRange:    [2]int{},
	}

	yearRange := f.yearRange.GetText()
	if yearRange != "" {
		splits := strings.Split(yearRange, "-")
		if len(splits) == 0 || len(splits) == 1 {
			if len(splits) == 1 {
				yearRange = splits[0]
			}
			firstYear, err := strconv.Atoi(yearRange)
			if err == nil {
				filt.YearRange[0] = firstYear
				filt.YearRange[1] = firstYear
			} else {
				logrus.Debugf("invalid year filter '%s': %v", yearRange, err)
			}
		} else if len(splits) == 2 {
			firstYear, err := strconv.Atoi(splits[0])
			secondYear, err := strconv.Atoi(splits[1])
			if err == nil {
				filt.YearRange[0] = firstYear
				filt.YearRange[1] = secondYear
			} else {
				logrus.Debugf("invalid year filter '%s': %v", yearRange, err)
			}
		}
	}
	f.filterFunc(filt)
	f.closeCb()
}

func (f *filter) InputHandler() func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p cview.Primitive)) {
		key := event.Key()
		if key == tcell.KeyEscape {
			f.closeCb()
		}
		f.Form.InputHandler()(event, setFocus)
	}
}
