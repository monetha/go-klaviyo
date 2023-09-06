package updater

// Profile is an interface that any profile update configuration should implement.
type Profile interface {
	Apply(map[string]interface{})
}

// ProfileFunc is a function type that implements the Profile interface.
type ProfileFunc func(map[string]interface{})

func (f ProfileFunc) Apply(profile map[string]interface{}) {
	f(profile)
}

// Location is an interface that any location update configuration should implement.
type Location interface {
	Apply(map[string]interface{})
}

// LocationFunc is a function type that implements the Location interface.
type LocationFunc func(map[string]interface{})

func (f LocationFunc) Apply(location map[string]interface{}) {
	f(location)
}

// Properties is an interface that any properties update configuration should implement.
type Properties interface {
	Apply(map[string]interface{})
}

// PropertiesFunc is a function type that implements the Properties interface.
type PropertiesFunc func(map[string]interface{})

func (f PropertiesFunc) Apply(properties map[string]interface{}) {
	f(properties)
}
