package profile

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// SchemaJSON returns a JSON Schema (draft 2020-12) describing a profile file.
// It is generated from the Profile type by reflection - field names from the
// json tags, constraints from the validate tags, allowed values from the
// exported value lists (RemovableApps and friends) - so it cannot drift from
// schema.go. It covers structure and enumerations; rules that need several
// fields at once (for example "custom_key requires product_key") are enforced
// by ValidateProfile only.
func SchemaJSON() ([]byte, error) {
	root := schemaFor(reflect.TypeOf(Profile{}), "")
	root["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	root["title"] = "unattend-gen profile"
	return json.MarshalIndent(root, "", "  ")
}

// enumValues maps the string types that have a closed list of values but no
// oneof tag to that list.
var enumValues = func() map[reflect.Type][]string {
	m := map[reflect.Type][]string{}
	add := func(list any) {
		v := reflect.ValueOf(list)
		vals := make([]string, v.Len())
		for i := range vals {
			vals[i] = v.Index(i).String()
		}
		m[v.Type().Elem()] = vals
	}
	add(RemovableApps)
	add(RemovableFeatures)
	add(RemovableOptionalFeatures)
	add(StickyKeysFlags)
	add(DesktopIcons)
	add(VisualEffects)
	add(StartFolders)
	add(VMGuestTools)
	return m
}()

type schemaMap = map[string]any

func schemaFor(t reflect.Type, validateTag string) schemaMap {
	rules := parseValidateTag(validateTag)

	if t.Kind() == reflect.Pointer {
		inner := schemaFor(t.Elem(), validateTag)
		return schemaMap{"anyOf": []any{inner, schemaMap{"type": "null"}}}
	}

	switch t.Kind() {
	case reflect.Struct:
		props := schemaMap{}
		var required []string
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" || !f.IsExported() {
				continue
			}
			props[name] = schemaFor(f.Type, f.Tag.Get("validate"))
			if _, ok := parseValidateTag(f.Tag.Get("validate"))["required"]; ok {
				required = append(required, name)
			}
		}
		s := schemaMap{"type": "object", "properties": props, "additionalProperties": false}
		if len(required) > 0 {
			s["required"] = required
		}
		return s

	case reflect.Slice:
		s := schemaMap{"type": "array", "items": schemaFor(t.Elem(), "")}
		if n, ok := rules["max"]; ok {
			if v, err := strconv.Atoi(n); err == nil {
				s["maxItems"] = v
			}
		}
		return s

	case reflect.Map:
		s := schemaMap{"type": "object", "additionalProperties": schemaFor(t.Elem(), "")}
		if vals, ok := enumValues[t.Key()]; ok {
			s["propertyNames"] = schemaMap{"enum": vals}
		}
		return s

	case reflect.String:
		s := schemaMap{"type": "string"}
		if one, ok := rules["oneof"]; ok {
			s["enum"] = strings.Fields(one)
		} else if vals, ok := enumValues[t]; ok {
			s["enum"] = vals
		}
		if n, ok := rules["max"]; ok {
			if v, err := strconv.Atoi(n); err == nil {
				s["maxLength"] = v
			}
		}
		return s

	case reflect.Bool:
		return schemaMap{"type": "boolean"}

	case reflect.Int, reflect.Int64:
		s := schemaMap{"type": "integer"}
		if n, ok := rules["eq"]; ok {
			if v, err := strconv.Atoi(n); err == nil {
				s["const"] = v
			}
		}
		return s
	}
	return schemaMap{}
}

// parseValidateTag splits a validate tag ("required,oneof=a b,max=5") into a
// map from rule name to parameter ("" for rules without one).
func parseValidateTag(tag string) map[string]string {
	rules := map[string]string{}
	for _, part := range strings.Split(tag, ",") {
		if part == "" {
			continue
		}
		name, param, _ := strings.Cut(part, "=")
		rules[name] = param
	}
	return rules
}
