package providers

import (
	"reflect"
	"strconv"
)

type ConfigField struct {
	Default any
	Key     string
	Label   string
	Type    string
}

func ConfigFields(p any) []ConfigField {
	rt := reflect.TypeOf(p)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}

	var fields []ConfigField
	for i := range rt.NumField() {
		f := rt.Field(i)
		tag := f.Tag.Get("config")
		if tag == "" {
			continue
		}
		fields = append(fields, ConfigField{
			Key:     tag,
			Label:   f.Tag.Get("label"),
			Type:    f.Tag.Get("type"),
			Default: f.Tag.Get("default"),
		})
	}
	return fields
}

func decodeConfig(dst any, cfg map[string]any) {
	rv := reflect.ValueOf(dst)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	rt := rv.Type()

	for i := range rt.NumField() {
		f := rt.Field(i)
		tag := f.Tag.Get("config")
		if tag == "" {
			continue
		}

		raw, ok := cfg[tag]
		if !ok {
			raw = f.Tag.Get("default")
		}
		if raw == "" {
			continue
		}

		fv := rv.Field(i)
		switch fv.Kind() {
		case reflect.String:
			s, _ := raw.(string)
			fv.SetString(s)
		case reflect.Int, reflect.Int64:
			var n int64
			switch v := raw.(type) {
			case float64:
				n = int64(v)
			case string:
				n, _ = strconv.ParseInt(v, 10, 64)
			}
			fv.SetInt(n)
		case reflect.Bool:
			var b bool
			switch v := raw.(type) {
			case bool:
				b = v
			case string:
				b, _ = strconv.ParseBool(v)
			}
			fv.SetBool(b)
		}
	}
}
