package updater

// ProfileData holds all the data needed to update the profile
type ProfileData struct {
	Attributes         map[string]interface{}
	PropertiesToRemove []string
}

// NewProfileData creates new instance of ProfileData
func NewProfileData() *ProfileData {
	return &ProfileData{
		Attributes: map[string]interface{}{},
	}
}

// Profile is an interface that any profile update configuration should implement.
type Profile interface {
	Apply(*ProfileData)
}

// ProfileFunc is a function type that implements the Profile interface.
type ProfileFunc func(*ProfileData)

func (f ProfileFunc) Apply(profile *ProfileData) {
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
