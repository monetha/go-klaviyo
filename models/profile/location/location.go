package location

import (
	"gitlab.com/monetha/go-klaviyo/models/profile/updater"
)

// WithAddress1 sets the Address1 for the location.
func WithAddress1(address1 string) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["address1"] = address1
	})
}

// WithAddress2 sets the Address2 for the location.
func WithAddress2(address2 string) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["address2"] = address2
	})
}

// WithCity sets the city for the location.
func WithCity(city string) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["city"] = city
	})
}

// WithCountry sets the country for the location.
func WithCountry(country string) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["country"] = country
	})
}

// WithLatitude sets the latitude for the location.
func WithLatitude(latitude float64) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["latitude"] = latitude
	})
}

// WithLongitude sets the longitude for the location.
func WithLongitude(longitude float64) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["longitude"] = longitude
	})
}

// WithRegion sets the region for the location.
func WithRegion(region string) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["region"] = region
	})
}

// WithZip sets the zip code for the location.
func WithZip(zip string) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["zip"] = zip
	})
}

// WithTimezone sets the timezone for the location.
func WithTimezone(timezone string) updater.Location {
	return updater.LocationFunc(func(location map[string]interface{}) {
		location["timezone"] = timezone
	})
}
