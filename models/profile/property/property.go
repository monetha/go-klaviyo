package property

import (
	"github.com/monetha/go-klaviyo/models/profile/updater"
)

// WithValue sets a specific key-value pair within the properties map.
//
// The name parameter specifies the key, and the value parameter specifies the value associated with that key.
func WithValue(name string, value interface{}) updater.Properties {
	return updater.PropertiesFunc(func(properties map[string]interface{}) {
		properties[name] = value
	})
}
