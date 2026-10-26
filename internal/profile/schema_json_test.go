package profile

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSchemaJSONDescribesProfile(t *testing.T) {
	data, err := SchemaJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s map[string]any
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if s["additionalProperties"] != false {
		t.Fatal("root must reject unknown properties")
	}
	props := s["properties"].(map[string]any)

	// Every json field of Profile is described.
	pt := reflect.TypeOf(Profile{})
	for i := 0; i < pt.NumField(); i++ {
		name := pt.Field(i).Tag.Get("json")
		if comma := indexByte(name, ','); comma >= 0 {
			name = name[:comma]
		}
		if _, ok := props[name]; !ok {
			t.Errorf("schema has no property %q", name)
		}
	}

	required := toStrings(s["required"].([]any))
	for _, want := range []string{"schema_version", "name"} {
		if !contains(required, want) {
			t.Errorf("%q should be required, got %v", want, required)
		}
	}

	edition := props["edition"].(map[string]any)["properties"].(map[string]any)["mode"].(map[string]any)
	if got := toStrings(edition["enum"].([]any)); !contains(got, "firmware") || !contains(got, "generic_key") {
		t.Errorf("edition.mode enum = %v", got)
	}

	apps := props["remove_apps"].(map[string]any)["items"].(map[string]any)
	if got := toStrings(apps["enum"].([]any)); len(got) != len(RemovableApps) {
		t.Errorf("remove_apps enum has %d values, want %d", len(got), len(RemovableApps))
	}
	if props["accounts"].(map[string]any)["maxItems"] != float64(5) {
		t.Errorf("accounts should have maxItems 5")
	}
	if props["schema_version"].(map[string]any)["const"] != float64(1) {
		t.Errorf("schema_version should be const 1")
	}
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func toStrings(in []any) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.(string)
	}
	return out
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
