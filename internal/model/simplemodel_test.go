package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/seaswalker/spring-analysis/internal/model"
)

// TestToString keeps the exact rendering the controller prints per request,
// including the four characters an absent field prints as.
func TestToString(t *testing.T) {
	m := model.NewSimpleModel()
	if got, want := m.String(), "SimpleModel{name='', age=null, date='null'}"; got != want {
		t.Errorf("empty model = %q, want %q", got, want)
	}

	age := 30
	moment := time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC)
	m.SetName("bob")
	m.SetAge(&age)
	m.SetDate(&moment)
	want := "SimpleModel{name='bob', age=30, date='Thu Jan 02 03:04:05 UTC 2020'}"
	if got := m.String(); got != want {
		t.Errorf("model = %q, want %q", got, want)
	}
}

// TestBindingAcceptsWhatJacksonAccepts covers the two date representations
// that bind, and the one that does not.
func TestBindingAcceptsWhatJacksonAccepts(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr bool
		want    string
	}{
		{"no date", `{"name":"bob","age":30}`, false, "SimpleModel{name='bob', age=30, date='null'}"},
		{"no age", `{"name":"bob"}`, false, "SimpleModel{name='bob', age=null, date='null'}"},
		{"epoch millis", `{"name":"bob","age":30,"date":1577934245000}`, false,
			"SimpleModel{name='bob', age=30, date='Thu Jan 02 03:04:05 UTC 2020'}"},
		{"iso 8601", `{"name":"bob","age":30,"date":"2020-01-02T03:04:05.000+0000"}`, false,
			"SimpleModel{name='bob', age=30, date='Thu Jan 02 03:04:05 UTC 2020'}"},
		{"rfc 3339", `{"name":"bob","age":30,"date":"2020-01-02T03:04:05Z"}`, false,
			"SimpleModel{name='bob', age=30, date='Thu Jan 02 03:04:05 UTC 2020'}"},
		// The @DateTimeFormat pattern configures the form data binder, not the
		// JSON converter, so a body carrying it is unreadable.
		{"DateTimeFormat pattern", `{"name":"bob","age":30,"date":"2020-01-02 03:04:05"}`, true, ""},
		{"not json", `nonsense`, true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := model.NewSimpleModel()
			err := json.Unmarshal([]byte(tc.body), m)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("binding %s succeeded, want a failure", tc.body)
				}
				return
			}
			if err != nil {
				t.Fatalf("binding %s: %v", tc.body, err)
			}
			if got := m.String(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestMarshalRoundTrip checks the model serialises back to a shape it can
// read.
func TestMarshalRoundTrip(t *testing.T) {
	age := 42
	moment := time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC)
	original := model.NewSimpleModel()
	original.SetName("bob")
	original.SetAge(&age)
	original.SetDate(&moment)

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	decoded := model.NewSimpleModel()
	if err := json.Unmarshal(encoded, decoded); err != nil {
		t.Fatalf("Unmarshal(%s): %v", encoded, err)
	}
	if decoded.String() != original.String() {
		t.Errorf("round trip changed the model: %q -> %q", original, decoded)
	}
}

// TestConstraints keeps the @Max bound and message the field declared.
func TestConstraints(t *testing.T) {
	constraints := model.NewSimpleModel().Constraints()
	if len(constraints) != 1 {
		t.Fatalf("declared %d constraints, want 1", len(constraints))
	}
	c := constraints[0]
	if c.Property != "age" || c.Message != model.AgeMaxMessage {
		t.Errorf("constraint = %+v", c)
	}
	ninety, ninetyOne := 90, 91
	if !c.Check(&ninety) {
		t.Error("an age at the bound was rejected")
	}
	if c.Check(&ninetyOne) {
		t.Error("an age past the bound was accepted")
	}
	if !c.Check(nil) {
		t.Error("a null age was rejected; JSR-380 treats null as valid")
	}
	if model.AgeMax != 90 {
		t.Errorf("AgeMax = %d, want 90", model.AgeMax)
	}
	if model.DateTimeFormatPattern != "2006-01-02 15:04:05" {
		t.Errorf("DateTimeFormatPattern = %q", model.DateTimeFormatPattern)
	}
}

// TestAccessors covers the bean surface the original declared.
func TestAccessors(t *testing.T) {
	m := model.NewSimpleModel()
	if m.GetName() != "" || m.GetAge() != nil || m.GetDate() != nil {
		t.Errorf("a fresh model is not empty: %v", m)
	}

	age := 21
	moment := time.Date(2021, time.March, 4, 5, 6, 7, 0, time.UTC)
	m.SetName("ann")
	m.SetAge(&age)
	m.SetDate(&moment)

	if m.GetName() != "ann" {
		t.Errorf("GetName = %q", m.GetName())
	}
	if m.GetAge() == nil || *m.GetAge() != 21 {
		t.Errorf("GetAge = %v", m.GetAge())
	}
	if m.GetDate() == nil || !m.GetDate().Equal(moment) {
		t.Errorf("GetDate = %v", m.GetDate())
	}
}
