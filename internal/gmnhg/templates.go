// This file is part of gmnhg.

// gmnhg is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gmnhg is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gmnhg. If not, see <https://www.gnu.org/licenses/>.

package gmnhg

import (
	"reflect"
	"sort"
)

func sortWhatever(sortable interface{}, reverse bool) interface{} {
	// convert slices to their sort.Interface counterparts
	switch s := sortable.(type) {
	case []int:
		sortable = sort.IntSlice(s)
	case []float64:
		sortable = sort.Float64Slice(s)
	case []string:
		sortable = sort.StringSlice(s)
	}
	v := reflect.ValueOf(sortable)
	cpy := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
	reflect.Copy(cpy, v)
	cpyAsInterface := v.Interface()
	if !reverse {
		sort.Sort(cpyAsInterface.(sort.Interface))
	} else {
		sort.Sort(sort.Reverse(cpyAsInterface.(sort.Interface)))
	}
	return cpyAsInterface
}

func Sort(sortable interface{}) interface{} {
	return sortWhatever(sortable, false)
}

func SortRev(sortable interface{}) interface{} {
	return sortWhatever(sortable, true)
}
