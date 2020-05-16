package jsonparse

import (
	"bytes"
	"encoding/json"
	"io"
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

	bodyCopy := bytes.Buffer{}
	tee := io.TeeReader(r.Body, &bodyCopy) // preserve the body
	err := json.NewDecoder(tee).Decode(&v)
	if err != nil {
		return nil, err
	}

	// replace the body for further handlers
	r.Body = ioutil.NopCloser(&bodyCopy)

	// prevent repetitive parsing. cache values
	values := map[string]interface{}{}

	return func(key string) (interface{}, bool) {
		prefix := "json."
		if !strings.HasPrefix(key, prefix) {
			return nil, false
		}
		key = strings.TrimPrefix(key, prefix)

		// use cache if previously fetched
		if val, ok := values[key]; ok {
			return val, true
		}

		val := fetchValue(v, key)
		values[key] = val // cache

		return val, true

	}, nil
}
