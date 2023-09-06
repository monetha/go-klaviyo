// Package provides utilities to define parameters for the GetProfiles method.

package getprofiles

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	minPageSize     = 1
	maxPageSize     = 100
	defaultPageSize = 20
)

// Param is an interface that any parameter type should implement.
// It provides a method to apply the parameter as a query parameter.
type Param interface {
	Apply(fields url.Values)
}

// FieldsUpdaterFunc is a type that wraps a function that updates URL query parameters.
type FieldsUpdaterFunc func(url.Values)

// Apply calls the underlying function to update the URL query parameters.
func (f FieldsUpdaterFunc) Apply(fields url.Values) {
	f(fields)
}

// WithDefaultPageSize returns a parameter that sets the page size to its default value.
func WithDefaultPageSize() Param {
	return WithPageSize(defaultPageSize)
}

// WithPageSize returns a parameter that sets the page size for the request.
// It ensures that the page size is within the allowed range.
func WithPageSize(pageSize int) Param {
	if pageSize < minPageSize {
		pageSize = minPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return FieldsUpdaterFunc(func(fields url.Values) {
		fields.Set("page[size]", strconv.Itoa(pageSize))
	})
}

// WithFields returns a parameter that sets the specific fields to be retrieved for the profile.
// It accepts a variable number of field names and constructs the appropriate query parameter.
func WithFields(fieldName ...string) Param {
	return FieldsUpdaterFunc(func(fields url.Values) {
		if names := strings.Join(fieldName, ","); names != "" {
			fields.Set("fields[profile]", names)
		}
	})
}
