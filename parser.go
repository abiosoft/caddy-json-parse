package jsonvars

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy/v2"
)

type fetcher interface {
	Fetch(interface{}, string) (interface{}, bool)
}

type fetcherFunc func(interface{}, string) (interface{}, bool)

func (f fetcherFunc) Fetch(v interface{}, key string) (interface{}, bool) {
	return f(v, key)
}

type fetchers []fetcher

func (fs fetchers) Fetch(v interface{}, key string) (interface{}, bool) {
	for _, f := range fs {
		v, ok := f.Fetch(v, key)
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
	f := fetchers{
		fetcherFunc(fromMap),
		fetcherFunc(fromArray),
	}

	var current interface{} = v
	for _, k := range strings.Split(key, ".") {
		val, ok := f.Fetch(current, k)
		if !ok {
			return nil
		}
		current = val
	}

	return current
}

func newReplacerFunc(r *http.Request) (caddy.ReplacerFunc, error) {
	var v interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}

	// prevent repetitive parsing. cache values
	keys := map[string]interface{}{}

	return func(key string) (interface{}, bool) {
		prefix := "json."
		if !strings.HasPrefix(key, prefix) {
			return nil, false
		}
		key = strings.TrimPrefix(key, prefix)

		// use cache if previously fetched
		if val, ok := keys[key]; ok {
			return val, true
		}

		return fetchValue(v, key), true

	}, nil
}
