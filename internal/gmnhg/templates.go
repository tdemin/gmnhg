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
