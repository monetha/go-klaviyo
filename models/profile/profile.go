package profile

import (
	"time"

	"gitlab.com/monetha/go-klaviyo/models/profile/location"
	"gitlab.com/monetha/go-klaviyo/models/profile/property"
	"gitlab.com/monetha/go-klaviyo/models/profile/updater"
)

// NewProfile represents the data structure for a profile that is not yet created.
type NewProfile struct {
	Attributes NewAttributes `json:"attributes"`
}

// ExistingProfile represents the data structure for a profile that is already created.
type ExistingProfile struct {
	Id         string             `json:"id"`
	Attributes ExistingAttributes `json:"attributes"`
}

// NewAttributes contains common attributes for a profile.
type NewAttributes struct {
	Email        string                 `json:"email"`
	PhoneNumber  *string                `json:"phone_number"`
	ExternalId   *string                `json:"external_id"`
	AnonymousId  *string                `json:"anonymous_id"`
	FirstName    *string                `json:"first_name"`
	LastName     *string                `json:"last_name"`
	Organization *string                `json:"organization"`
	Title        *string                `json:"title"`
	Image        *string                `json:"image"`
	Location     Location               `json:"location"`
	Properties   map[string]interface{} `json:"properties"`
}

// ExistingAttributes contains attributes for a profile that is already created, including timestamps.
type ExistingAttributes struct {
	NewAttributes
	Created       time.Time  `json:"created"`
	Updated       time.Time  `json:"updated"`
	LastEventDate *time.Time `json:"last_event_date"`
}

// Location represents the geographical location details for a profile.
type Location struct {
	Address1  *string  `json:"address1"`
	Address2  *string  `json:"address2"`
	City      *string  `json:"city"`
	Country   *string  `json:"country"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Region    *string  `json:"region"`
	Zip       *string  `json:"zip"`
	Timezone  *string  `json:"timezone"`
}

// WithEmail sets the email for the profile.
func WithEmail(email string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["email"] = email
	})
}

// WithPhoneNumber sets the phone number for the profile.
func WithPhoneNumber(phoneNumber string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["phone_number"] = phoneNumber
	})
}

// WithExternalId sets the external ID for the profile.
func WithExternalId(externalId string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["external_id"] = externalId
	})
}

// WithAnonymousId sets the anonymous ID for the profile.
func WithAnonymousId(anonymousId string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["anonymous_id"] = anonymousId
	})
}

// WithFirstName sets the first name for the profile.
func WithFirstName(firstName string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["first_name"] = firstName
	})
}

// WithLastName sets the last name for the profile.
func WithLastName(lastName string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["last_name"] = lastName
	})
}

// WithOrganization sets the organization for the profile.
func WithOrganization(organization string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["organization"] = organization
	})
}

// WithTitle sets the title for the profile.
func WithTitle(title string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["title"] = title
	})
}

// WithImage sets the image URL for the profile.
func WithImage(image string) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		profile["image"] = image
	})
}

// WithLocation sets the location for the profile.
func WithLocation(updaters ...updater.Location) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		loc := make(map[string]interface{})
		for _, u := range updaters {
			u.Apply(loc)
		}
		profile["location"] = loc
	})
}

// WithProperties sets the properties for the profile.
//
// It accepts a variable number of updaters that each set a specific property.
// Each updater is responsible for setting a specific key-value pair within the properties map.
func WithProperties(updaters ...updater.Properties) updater.Profile {
	return updater.ProfileFunc(func(profile map[string]interface{}) {
		properties := make(map[string]interface{})
		for _, u := range updaters {
			u.Apply(properties)
		}
		profile["properties"] = properties
	})
}

// ToUpdaters takes a NewProfile and transforms it into a slice of updater.Profile.
// This function facilitates the conversion of a profile's fields into a series of updaters,
// which can be used to modify a profile in a more granular manner. Importantly, it creates updaters
// only for the non-nil fields of the profile, ensuring that only specified fields are updated.
// It handles all the fields of the profile, including nested fields like Location and Properties.
func (p *NewProfile) ToUpdaters() []updater.Profile {
	if p == nil {
		return nil
	}

	var updaters []updater.Profile

	// NewAttributes
	attr := p.Attributes

	// Email
	if attr.Email != "" {
		updaters = append(updaters, WithEmail(attr.Email))
	}

	// PhoneNumber
	if attr.PhoneNumber != nil {
		updaters = append(updaters, WithPhoneNumber(*attr.PhoneNumber))
	}

	// ExternalId
	if attr.ExternalId != nil {
		updaters = append(updaters, WithExternalId(*attr.ExternalId))
	}

	// AnonymousId
	if attr.AnonymousId != nil {
		updaters = append(updaters, WithAnonymousId(*attr.AnonymousId))
	}

	// FirstName
	if attr.FirstName != nil {
		updaters = append(updaters, WithFirstName(*attr.FirstName))
	}

	// LastName
	if attr.LastName != nil {
		updaters = append(updaters, WithLastName(*attr.LastName))
	}

	// Organization
	if attr.Organization != nil {
		updaters = append(updaters, WithOrganization(*attr.Organization))
	}

	// Title
	if attr.Title != nil {
		updaters = append(updaters, WithTitle(*attr.Title))
	}

	// Image
	if attr.Image != nil {
		updaters = append(updaters, WithImage(*attr.Image))
	}

	// Location
	loc := attr.Location
	var locationUpdaters []updater.Location
	if loc.Address1 != nil {
		locationUpdaters = append(locationUpdaters, location.WithAddress1(*loc.Address1))
	}
	if loc.Address2 != nil {
		locationUpdaters = append(locationUpdaters, location.WithAddress2(*loc.Address2))
	}
	if loc.City != nil {
		locationUpdaters = append(locationUpdaters, location.WithCity(*loc.City))
	}
	if loc.Country != nil {
		locationUpdaters = append(locationUpdaters, location.WithCountry(*loc.Country))
	}
	if loc.Latitude != nil {
		locationUpdaters = append(locationUpdaters, location.WithLatitude(*loc.Latitude))
	}
	if loc.Longitude != nil {
		locationUpdaters = append(locationUpdaters, location.WithLongitude(*loc.Longitude))
	}
	if loc.Region != nil {
		locationUpdaters = append(locationUpdaters, location.WithRegion(*loc.Region))
	}
	if loc.Zip != nil {
		locationUpdaters = append(locationUpdaters, location.WithZip(*loc.Zip))
	}
	if loc.Timezone != nil {
		locationUpdaters = append(locationUpdaters, location.WithTimezone(*loc.Timezone))
	}
	if len(locationUpdaters) > 0 {
		updaters = append(updaters, WithLocation(locationUpdaters...))
	}

	// Properties
	var propertiesUpdaters []updater.Properties
	for key, value := range attr.Properties {
		propertiesUpdaters = append(propertiesUpdaters, property.WithValue(key, value))
	}
	if len(propertiesUpdaters) > 0 {
		updaters = append(updaters, WithProperties(propertiesUpdaters...))
	}

	return updaters
}
