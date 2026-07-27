package providers

import "reflect"

type ConfigField struct {
	Default any
	Key     string
	Label   string
	Type    reflect.Kind
}

func ConfigFields(p any) []ConfigField {
	var fields []ConfigField
	forEachField(p, func(tag string, f reflect.StructField, _ reflect.Value) {
		fields = append(fields, ConfigField{
			Key:     tag,
			Label:   f.Tag.Get("label"),
			Type:    f.Type.Kind(),
			Default: f.Tag.Get("default"),
		})
	})
	return fields
}

func decodeConfig(dst any, cfg map[string]any) {
	forEachField(dst, func(tag string, f reflect.StructField, v reflect.Value) {
		raw, ok := cfg[tag]
		if !ok {
			raw = f.Tag.Get("default")
		}
		if s, ok := raw.(string); ok && s != "" {
			v.SetString(s)
		}
	})
}

func forEachField(p any, fn func(tag string, f reflect.StructField, v reflect.Value)) {
	rv := reflect.ValueOf(p)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	for f, v := range rv.Fields() {
		tag := f.Tag.Get("config")
		if tag == "" {
			continue
		}
		fn(tag, f, v)
	}
}
