package jsonvars

import (
	"strconv"
	"strings"
)

type fetcherFunc func(interface{}, string) (interface{}, bool)
type fetcherFuncs []fetcherFunc

func (fs fetcherFuncs) fetch(v interface{}, key string) (interface{}, bool) {
	for _, f := range fs {
		v, ok := f(v, key)
		if ok {
			return v, true
		}
	}
	return nil, false
}

func fromMap(v interface{}, key string) (interface{}, bool) {
	// convert value to map
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, false
	}

	// ensure key exists
	if val, ok := m[key]; ok {
		return val, true
	}
	return nil, true
}

func fromArray(v interface{}, key string) (interface{}, bool) {
	// convert key to int
	i, err := strconv.Atoi(key)
	if err != nil {
		return nil, false
	}

	// convert value to array
	a, ok := v.([]interface{})
	if !ok {
		return nil, false
	}

	// ensure index
	if len(a) > i {
		return a[i], true
	}

	return nil, true
}

func fetchValue(v interface{}, key string) interface{} {
	funcs := fetcherFuncs{fromMap, fromArray}
	var current interface{} = v
	for _, k := range strings.Split(key, ".") {
		val, ok := funcs.fetch(current, k)
		if !ok {
			return nil
		}
		current = val
	}
	return current
}
