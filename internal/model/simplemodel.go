// Package model holds the model bound from the web application's request
// bodies.
package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
)

// DateTimeFormatPattern is the pattern the original declares with
// @DateTimeFormat(pattern = "yyyy-MM-dd HH:mm:ss").
//
// It configures Spring's DataBinder, not the JSON message converter, so a
// request body carrying a date in this shape is *not* parsed by it — the
// binding fails, exactly as it does in the original. The pattern is kept
// because it is part of the declared model, and because form binding would
// use it.
const DateTimeFormatPattern = "2006-01-02 15:04:05"

// AgeMaxMessage is the message @Max(value = 90) declares.
const AgeMaxMessage = "年龄最大不能超过90"

// AgeMax is the bound @Max declares.
const AgeMax = 90

// SimpleModel is the request model. Age is a pointer because the original's
// field is a boxed Integer, and the difference between absent and zero is
// printed.
type SimpleModel struct {
	name string
	age  *int
	date *time.Time
}

// NewSimpleModel returns an empty model.
func NewSimpleModel() *SimpleModel { return &SimpleModel{} }

// GetName returns the name.
func (m *SimpleModel) GetName() string { return m.name }

// SetName sets the name.
func (m *SimpleModel) SetName(name string) { m.name = name }

// GetAge returns the age, or nil when the request omitted it.
func (m *SimpleModel) GetAge() *int { return m.age }

// SetAge sets the age.
func (m *SimpleModel) SetAge(age *int) { m.age = age }

// GetDate returns the date, or nil when the request omitted it.
func (m *SimpleModel) GetDate() *time.Time { return m.date }

// SetDate sets the date.
func (m *SimpleModel) SetDate(date *time.Time) { m.date = date }

// Constraints declares the JSR-380 constraints the original's field
// annotations carry.
func (m *SimpleModel) Constraints() []validation.Constraint {
	return []validation.Constraint{
		validation.Max("age", AgeMax, AgeMaxMessage),
	}
}

// String reproduces SimpleModel.toString() exactly, including the quoting of
// name and date but not of age, and "null" for an absent value.
func (m *SimpleModel) String() string {
	return fmt.Sprintf("SimpleModel{name='%s', age=%s, date='%s'}",
		m.name, javart.ToString(m.age), javart.DateToString(m.date))
}

// wire is the JSON shape the message converter reads, matching what Jackson
// accepts for these fields.
type wire struct {
	Name string           `json:"name"`
	Age  *int             `json:"age"`
	Date *json.RawMessage `json:"date"`
}

// UnmarshalJSON binds a request body, reproducing what Jackson does with a
// java.util.Date field: epoch milliseconds and ISO-8601 parse, and anything
// else — including the @DateTimeFormat pattern — is a binding failure.
func (m *SimpleModel) UnmarshalJSON(data []byte) error {
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	m.name, m.age, m.date = w.Name, w.Age, nil
	if w.Date == nil {
		return nil
	}
	date, err := parseJacksonDate(*w.Date)
	if err != nil {
		return err
	}
	m.date = date
	return nil
}

// MarshalJSON writes the shape Jackson would serialise, so a round trip is
// symmetric.
func (m *SimpleModel) MarshalJSON() ([]byte, error) {
	out := struct {
		Name string `json:"name"`
		Age  *int   `json:"age"`
		Date *int64 `json:"date"`
	}{Name: m.name, Age: m.age}
	if m.date != nil {
		millis := m.date.UnixMilli()
		out.Date = &millis
	}
	return json.Marshal(out)
}

// parseJacksonDate accepts the two representations Jackson binds to a
// java.util.Date by default: epoch milliseconds as a number, and an ISO-8601
// timestamp as a string.
func parseJacksonDate(raw json.RawMessage) (*time.Time, error) {
	var millis int64
	if err := json.Unmarshal(raw, &millis); err == nil {
		t := time.UnixMilli(millis).UTC()
		return &t, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return nil, fmt.Errorf("cannot deserialize value of type `java.util.Date` from %s", raw)
	}
	for _, layout := range []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z0700",
		time.RFC3339Nano,
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, text); err == nil {
			return ptr(t.UTC()), nil
		}
	}
	return nil, fmt.Errorf("cannot deserialize value of type `java.util.Date` from String %q: "+
		"not a valid representation (error: Failed to parse Date value '%s')", text, text)
}

func ptr[T any](v T) *T { return &v }

var _ validation.Constrained = (*SimpleModel)(nil)
