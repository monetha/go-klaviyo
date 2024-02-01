package klaviyo_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/monetha/go-klaviyo/models/event"
	"io"
	"net/http"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/monetha/go-klaviyo"
	"github.com/monetha/go-klaviyo/models/profile"
	"github.com/monetha/go-klaviyo/models/profile/property"
	"github.com/monetha/go-klaviyo/operations/getprofiles"
)

const (
	validAPIKey   = "valid-api-key"
	invalidAPIKey = "invalid-api-key"
)

func TestClient_GetProfiles(t *testing.T) {
	t.Run("get profiles with invalid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_profiles_invalid_api_key", func(c *http.Client) {
			kc := klaviyo.NewWithClient(invalidAPIKey, zap.L(), c)

			ctx := context.TODO()
			ps, err := kc.GetProfiles(ctx)

			require.ErrorIs(t, err, klaviyo.ErrInvalidAPIKey)
			require.Nil(t, ps)
		})
	})

	t.Run("get profiles with correctly formatted but invalid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_profiles_correctly_formatted_invalid_api_key", func(c *http.Client) {
			kc := klaviyo.NewWithClient("pk_1111111111111111111111111111111112", zap.L(), c)

			ctx := context.TODO()
			ps, err := kc.GetProfiles(ctx)

			require.ErrorIs(t, err, klaviyo.ErrInvalidAPIKey)
			require.Nil(t, ps)
		})
	})

	t.Run("get profiles with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_profiles_valid_api_key", func(c *http.Client) {
			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			ps, err := kc.GetProfiles(ctx)

			require.NoError(t, err)
			require.Len(t, ps, 2, "profiles len")
		})
	})

	t.Run("get profiles with email and phone using valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_profiles_with_email_and_phone_valid_api_key", func(c *http.Client) {
			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			ps, err := kc.GetProfiles(ctx,
				getprofiles.WithPageSize(3),
				getprofiles.WithFields("email", "phone_number"),
			)

			require.NoError(t, err)
			require.Len(t, ps, 3, "profiles len")
		})
	})
}

var initialProfile = &profile.NewProfile{
	Attributes: profile.NewAttributes{
		Email:        "sarah.mason@klaviyo-demo.com",
		PhoneNumber:  pVal("+15005550006"),
		ExternalId:   pVal("63f64a2b-c6bf-40c7-b81f-bed08162edbe"),
		AnonymousId:  pVal("anon-63f64a2b-c6bf-40c7-b81f-bed08162edbe"),
		FirstName:    pVal("Sarah"),
		LastName:     pVal("Mason"),
		Organization: pVal("Klaviyo"),
		Title:        pVal("Engineer"),
		Image:        pVal("https://images.pexels.com/photos/3760854/pexels-photo-3760854.jpeg"),
		Location: profile.Location{
			Address1:  pVal("89 E 42nd St"),
			Address2:  pVal("1st floor"),
			City:      pVal("New York"),
			Country:   pVal("United States"),
			Latitude:  pVal(56.0),
			Longitude: pVal(24.0),
			Region:    pVal("NY"),
			Zip:       pVal("10017"),
			Timezone:  pVal("America/New_York"),
		},
		Properties: map[string]interface{}{
			"pseudonym": "Dr. Octopus",
		},
	},
}

var inititalEvent = event.NewEvent{
	NewAttributes: event.NewAttributes{
		Time:  "2024-01-30T05:10:00",
		Value: 0,
		Properties: map[string]string{
			"EventName":    "EmailSent",
			"PointClaimed": "1500",
			"PointOverall": "20000",
		},
	},
}

func TestClient_CreateProfile(t *testing.T) {
	t.Run("create profile with invalid API key", func(t *testing.T) {
		withHTTPRecorder("tests/create_profile_invalid_api_key", func(c *http.Client) {
			kc := klaviyo.NewWithClient(invalidAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.CreateProfile(ctx, initialProfile)

			require.ErrorIs(t, err, klaviyo.ErrInvalidAPIKey)
			require.Nil(t, cp)
		})
	})

	t.Run("create profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/create_profile_valid_api_key", func(c *http.Client) {
			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()

			cp, err := kc.CreateProfile(ctx, initialProfile)

			require.NoError(t, err)
			require.NotNil(t, cp)

			// Additional checks to ensure created profile has the same values
			require.Equal(t, "01H8HKMDG8F4MN7PSRZ4YQYNVQ", cp.Id, "Mismatch in field: Id")
			initialProfAttrs := initialProfile.Attributes
			profAttrs := cp.Attributes
			require.Equal(t, initialProfAttrs.Email, profAttrs.Email, "Mismatch in field: Email")
			require.Equal(t, initialProfAttrs.PhoneNumber, profAttrs.PhoneNumber, "Mismatch in field: PhoneNumber")
			require.Equal(t, initialProfAttrs.ExternalId, profAttrs.ExternalId, "Mismatch in field: ExternalId")
			require.Equal(t, initialProfAttrs.AnonymousId, profAttrs.AnonymousId, "Mismatch in field: AnonymousId")
			require.Equal(t, initialProfAttrs.FirstName, profAttrs.FirstName, "Mismatch in field: FirstName")
			require.Equal(t, initialProfAttrs.LastName, profAttrs.LastName, "Mismatch in field: LastName")
			require.Equal(t, initialProfAttrs.Organization, profAttrs.Organization, "Mismatch in field: Organization")
			require.Equal(t, initialProfAttrs.Title, profAttrs.Title, "Mismatch in field: Title")
			require.Equal(t, initialProfAttrs.Image, profAttrs.Image, "Mismatch in field: Image")
			require.Equal(t, initialProfAttrs.Location, profAttrs.Location, "Mismatch in field: Location")
			require.Equal(t, initialProfAttrs.Properties, profAttrs.Properties, "Mismatch in field: Properties")
		})
	})

	t.Run("create existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/create_existing_profile_valid_api_key", func(c *http.Client) {
			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()

			cp, err := kc.CreateProfile(ctx, initialProfile)

			require.NotNil(t, err)
			require.Nil(t, cp)

			// Type assert error to *ErrProfileAlreadyExists and compare the DuplicateProfileID
			var e *klaviyo.ErrProfileAlreadyExists
			if errors.As(err, &e) {
				require.Equal(t, "01H8HKMDG8F4MN7PSRZ4YQYNVQ", e.DuplicateProfileID)
			} else {
				t.Fatalf("expected error of type *klaviyo.ErrProfileAlreadyExists, got %T", err)
			}
		})
	})
}

func TestClient_GetProfile(t *testing.T) {
	t.Run("get existing profile with invalid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_existing_profile_invalid_api_key", func(c *http.Client) {
			const existingProfileID = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"

			kc := klaviyo.NewWithClient(invalidAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.GetProfile(ctx, existingProfileID)

			require.ErrorIs(t, err, klaviyo.ErrInvalidAPIKey)
			require.Nil(t, cp)
		})
	})

	t.Run("get existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_existing_profile_valid_api_key", func(c *http.Client) {
			const existingProfileID = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.GetProfile(ctx, existingProfileID)

			require.NoError(t, err)
			require.NotNil(t, cp)

			// Additional checks to ensure created profile has the same values
			require.Equal(t, existingProfileID, cp.Id, "Mismatch in field: Id")
			initialProfAttrs := initialProfile.Attributes
			profAttrs := cp.Attributes
			require.Equal(t, initialProfAttrs.Email, profAttrs.Email, "Mismatch in field: Email")
			require.Equal(t, initialProfAttrs.PhoneNumber, profAttrs.PhoneNumber, "Mismatch in field: PhoneNumber")
			require.Equal(t, initialProfAttrs.ExternalId, profAttrs.ExternalId, "Mismatch in field: ExternalId")
			require.Equal(t, initialProfAttrs.AnonymousId, profAttrs.AnonymousId, "Mismatch in field: AnonymousId")
			require.Equal(t, initialProfAttrs.FirstName, profAttrs.FirstName, "Mismatch in field: FirstName")
			require.Equal(t, initialProfAttrs.LastName, profAttrs.LastName, "Mismatch in field: LastName")
			require.Equal(t, initialProfAttrs.Organization, profAttrs.Organization, "Mismatch in field: Organization")
			require.Equal(t, initialProfAttrs.Title, profAttrs.Title, "Mismatch in field: Title")
			require.Equal(t, initialProfAttrs.Image, profAttrs.Image, "Mismatch in field: Image")
			require.Equal(t, initialProfAttrs.Location, profAttrs.Location, "Mismatch in field: Location")
			require.Equal(t, initialProfAttrs.Properties, profAttrs.Properties, "Mismatch in field: Properties")
		})
	})

	t.Run("get non-existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_non_existing_profile_valid_api_key", func(c *http.Client) {
			const nonExistingProfileID = "UQHWDB2XIYWHF9GYUWCY04KU8O"

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.GetProfile(ctx, nonExistingProfileID)

			require.ErrorIs(t, err, klaviyo.ErrProfileDoesNotExist)
			require.Nil(t, cp)
		})
	})
}

func TestClient_UpdateProfile(t *testing.T) {
	t.Run("update existing profile with invalid API key", func(t *testing.T) {
		withHTTPRecorder("tests/update_existing_profile_invalid_api_key", func(c *http.Client) {
			const existingProfileID = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"

			kc := klaviyo.NewWithClient(invalidAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.UpdateProfile(ctx,
				existingProfileID,
				initialProfile.ToUpdaters()...)

			require.ErrorIs(t, err, klaviyo.ErrInvalidAPIKey)
			require.Nil(t, cp)
		})
	})

	t.Run("update existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/update_existing_profile_valid_api_key", func(c *http.Client) {
			const existingProfileID = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.UpdateProfile(ctx,
				existingProfileID,
				initialProfile.ToUpdaters()...)

			require.NoError(t, err)
			require.NotNil(t, cp)

			// Additional checks to ensure created profile has the same values
			require.Equal(t, existingProfileID, cp.Id, "Mismatch in field: Id")
			initialProfAttrs := initialProfile.Attributes
			profAttrs := cp.Attributes
			require.Equal(t, initialProfAttrs.Email, profAttrs.Email, "Mismatch in field: Email")
			require.Equal(t, initialProfAttrs.PhoneNumber, profAttrs.PhoneNumber, "Mismatch in field: PhoneNumber")
			require.Equal(t, initialProfAttrs.ExternalId, profAttrs.ExternalId, "Mismatch in field: ExternalId")
			require.Equal(t, initialProfAttrs.AnonymousId, profAttrs.AnonymousId, "Mismatch in field: AnonymousId")
			require.Equal(t, initialProfAttrs.FirstName, profAttrs.FirstName, "Mismatch in field: FirstName")
			require.Equal(t, initialProfAttrs.LastName, profAttrs.LastName, "Mismatch in field: LastName")
			require.Equal(t, initialProfAttrs.Organization, profAttrs.Organization, "Mismatch in field: Organization")
			require.Equal(t, initialProfAttrs.Title, profAttrs.Title, "Mismatch in field: Title")
			require.Equal(t, initialProfAttrs.Image, profAttrs.Image, "Mismatch in field: Image")
			require.Equal(t, initialProfAttrs.Location, profAttrs.Location, "Mismatch in field: Location")
			require.Equal(t, initialProfAttrs.Properties, profAttrs.Properties, "Mismatch in field: Properties")
		})
	})

	t.Run("update phone only for the existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/update_phone_existing_profile_valid_api_key", func(c *http.Client) {
			const (
				existingProfileID = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"
				newPhoneNumber    = "+15005550007"
			)

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.UpdateProfile(ctx,
				existingProfileID,
				profile.WithPhoneNumber(newPhoneNumber),
			)

			require.NoError(t, err)
			require.NotNil(t, cp)

			// Additional checks to ensure created profile has the same values
			require.Equal(t, existingProfileID, cp.Id, "Mismatch in field: Id")
			initialProfAttrs := initialProfile.Attributes
			profAttrs := cp.Attributes
			require.Equal(t, initialProfAttrs.Email, profAttrs.Email, "Mismatch in field: Email")
			require.Equal(t, pVal(newPhoneNumber), profAttrs.PhoneNumber, "Mismatch in field: PhoneNumber")
			require.Equal(t, initialProfAttrs.ExternalId, profAttrs.ExternalId, "Mismatch in field: ExternalId")
			require.Equal(t, initialProfAttrs.AnonymousId, profAttrs.AnonymousId, "Mismatch in field: AnonymousId")
			require.Equal(t, initialProfAttrs.FirstName, profAttrs.FirstName, "Mismatch in field: FirstName")
			require.Equal(t, initialProfAttrs.LastName, profAttrs.LastName, "Mismatch in field: LastName")
			require.Equal(t, initialProfAttrs.Organization, profAttrs.Organization, "Mismatch in field: Organization")
			require.Equal(t, initialProfAttrs.Title, profAttrs.Title, "Mismatch in field: Title")
			require.Equal(t, initialProfAttrs.Image, profAttrs.Image, "Mismatch in field: Image")
			require.Equal(t, initialProfAttrs.Location, profAttrs.Location, "Mismatch in field: Location")
			require.Equal(t, initialProfAttrs.Properties, profAttrs.Properties, "Mismatch in field: Properties")
		})
	})

	t.Run("update property only for the existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/update_property_existing_profile_valid_api_key", func(c *http.Client) {
			const (
				existingProfileID = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"
				newPseudonym      = "Ms. Octopus"
			)

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()

			cp, err := kc.UpdateProfile(ctx,
				existingProfileID,
				profile.WithProperties(
					property.WithValue("pseudonym", newPseudonym),
				),
			)

			require.NoError(t, err)
			require.NotNil(t, cp)

			// Additional checks to ensure created profile has the same values
			require.Equal(t, existingProfileID, cp.Id, "Mismatch in field: Id")
			initialProfAttrs := initialProfile.Attributes
			profAttrs := cp.Attributes
			require.Equal(t, initialProfAttrs.Email, profAttrs.Email, "Mismatch in field: Email")
			require.Equal(t, initialProfAttrs.PhoneNumber, profAttrs.PhoneNumber, "Mismatch in field: PhoneNumber")
			require.Equal(t, initialProfAttrs.ExternalId, profAttrs.ExternalId, "Mismatch in field: ExternalId")
			require.Equal(t, initialProfAttrs.AnonymousId, profAttrs.AnonymousId, "Mismatch in field: AnonymousId")
			require.Equal(t, initialProfAttrs.FirstName, profAttrs.FirstName, "Mismatch in field: FirstName")
			require.Equal(t, initialProfAttrs.LastName, profAttrs.LastName, "Mismatch in field: LastName")
			require.Equal(t, initialProfAttrs.Organization, profAttrs.Organization, "Mismatch in field: Organization")
			require.Equal(t, initialProfAttrs.Title, profAttrs.Title, "Mismatch in field: Title")
			require.Equal(t, initialProfAttrs.Image, profAttrs.Image, "Mismatch in field: Image")
			require.Equal(t, initialProfAttrs.Location, profAttrs.Location, "Mismatch in field: Location")
			require.Equal(t, map[string]interface{}{"pseudonym": newPseudonym}, profAttrs.Properties, "Mismatch in field: Properties")
		})
	})

	t.Run("update with new property for the existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/update_new_property_existing_profile_valid_api_key", func(c *http.Client) {
			const (
				existingProfileID = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"
				newPropertyName   = "skype"
				newPropertyValue  = "sarah_mason_skype"
			)

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()

			cp, err := kc.UpdateProfile(ctx,
				existingProfileID,
				profile.WithProperties(
					property.WithValue(newPropertyName, newPropertyValue),
				),
			)

			require.NoError(t, err)
			require.NotNil(t, cp)

			// Additional checks to ensure created profile has the same values
			require.Equal(t, existingProfileID, cp.Id, "Mismatch in field: Id")
			initialProfAttrs := initialProfile.Attributes
			clonedInitialProperties := cloneMap(initialProfAttrs.Properties)
			clonedInitialProperties[newPropertyName] = newPropertyValue
			profAttrs := cp.Attributes
			require.Equal(t, initialProfAttrs.Email, profAttrs.Email, "Mismatch in field: Email")
			require.Equal(t, initialProfAttrs.PhoneNumber, profAttrs.PhoneNumber, "Mismatch in field: PhoneNumber")
			require.Equal(t, initialProfAttrs.ExternalId, profAttrs.ExternalId, "Mismatch in field: ExternalId")
			require.Equal(t, initialProfAttrs.AnonymousId, profAttrs.AnonymousId, "Mismatch in field: AnonymousId")
			require.Equal(t, initialProfAttrs.FirstName, profAttrs.FirstName, "Mismatch in field: FirstName")
			require.Equal(t, initialProfAttrs.LastName, profAttrs.LastName, "Mismatch in field: LastName")
			require.Equal(t, initialProfAttrs.Organization, profAttrs.Organization, "Mismatch in field: Organization")
			require.Equal(t, initialProfAttrs.Title, profAttrs.Title, "Mismatch in field: Title")
			require.Equal(t, initialProfAttrs.Image, profAttrs.Image, "Mismatch in field: Image")
			require.Equal(t, initialProfAttrs.Location, profAttrs.Location, "Mismatch in field: Location")
			require.Equal(t, clonedInitialProperties, profAttrs.Properties, "Mismatch in field: Properties")
		})
	})

	t.Run("unset property for the existing profile with valid API KEY", func(t *testing.T) {
		withHTTPRecorder("tests/update_unset_property_existing_profile_valid_api_key", func(c *http.Client) {
			const (
				existingProfileID     = "01H8HKMDG8F4MN7PSRZ4YQYNVQ"
				skypePropertyName     = "skype"
				pseudonymPropertyName = "pseudonym"
			)

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()

			cp, err := kc.UpdateProfile(ctx,
				existingProfileID,
				profile.UnsetProperties(
					skypePropertyName, pseudonymPropertyName,
				),
			)

			require.NoError(t, err)
			require.NotNil(t, cp)

			// Additional checks to ensure created profile has the same values
			require.Equal(t, existingProfileID, cp.Id, "Mismatch in field: Id")
			initialProfAttrs := initialProfile.Attributes
			clonedInitialProperties := cloneMap(initialProfAttrs.Properties)
			delete(clonedInitialProperties, skypePropertyName)
			delete(clonedInitialProperties, pseudonymPropertyName)
			profAttrs := cp.Attributes
			require.Equal(t, initialProfAttrs.Email, profAttrs.Email, "Mismatch in field: Email")
			require.Equal(t, initialProfAttrs.PhoneNumber, profAttrs.PhoneNumber, "Mismatch in field: PhoneNumber")
			require.Equal(t, initialProfAttrs.ExternalId, profAttrs.ExternalId, "Mismatch in field: ExternalId")
			require.Equal(t, initialProfAttrs.AnonymousId, profAttrs.AnonymousId, "Mismatch in field: AnonymousId")
			require.Equal(t, initialProfAttrs.FirstName, profAttrs.FirstName, "Mismatch in field: FirstName")
			require.Equal(t, initialProfAttrs.LastName, profAttrs.LastName, "Mismatch in field: LastName")
			require.Equal(t, initialProfAttrs.Organization, profAttrs.Organization, "Mismatch in field: Organization")
			require.Equal(t, initialProfAttrs.Title, profAttrs.Title, "Mismatch in field: Title")
			require.Equal(t, initialProfAttrs.Image, profAttrs.Image, "Mismatch in field: Image")
			require.Equal(t, initialProfAttrs.Location, profAttrs.Location, "Mismatch in field: Location")
			require.Equal(t, clonedInitialProperties, profAttrs.Properties, "Mismatch in field: Properties")
		})
	})

	t.Run("update non-existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/update_non_existing_profile_valid_api_key", func(c *http.Client) {
			const nonExistingProfileID = "UQHWDB2XIYWHF9GYUWCY04KU8O"

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			cp, err := kc.UpdateProfile(ctx,
				nonExistingProfileID,
				initialProfile.ToUpdaters()...)

			require.ErrorIs(t, err, klaviyo.ErrProfileDoesNotExist)
			require.Nil(t, cp)
		})
	})
}

func TestClient_Events(t *testing.T) {
	t.Run("create new event with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/create_new_event_valid_api_key", func(c *http.Client) {
			const existingProfileID = "01HN6AFEHGF6F77WJRKT1C9JHG"

			metricName := "Reward"

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			err := kc.CreateEvent(ctx, &inititalEvent, existingProfileID, metricName)

			require.NoError(t, err)
		})
	})

	t.Run("get existing profile with valid API key", func(t *testing.T) {
		withHTTPRecorder("tests/get_existing_event_valid_api_key", func(c *http.Client) {

			kc := klaviyo.NewWithClient(validAPIKey, zap.L(), c)

			ctx := context.TODO()
			ce, err := kc.GetEvents(ctx)

			require.NoError(t, err)
			require.NotNil(t, ce)
			require.Len(t, ce, 1)

			result := ce[0]
			prop := result.Attributes.EventProperties

			require.Equal(t, result.Attributes.UUID, "d13e0400-bf2d-11ee-8001-dd51f1217edd")
			require.NotEmpty(t, prop)
			require.Equal(t, prop["EventName"], inititalEvent.Properties["EventName"])
			require.Equal(t, prop["PointClaimed"], inititalEvent.Properties["PointClaimed"])
			require.Equal(t, prop["PointOverall"], inititalEvent.Properties["PointOverall"])
		})
	})
}

func pVal[T any](val T) *T { return &val }

func cloneMap[M ~map[K]V, K comparable, V any](src M) M {
	if src == nil {
		return nil
	}
	dst := make(M)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func withHTTPRecorder(cassetteName string, f func(*http.Client)) {
	// start http recorder
	r, err := recorder.New(cassetteName)
	if err != nil {
		panic(err)
	}
	defer func() { _ = r.Stop() }() // Make sure recorder is stopped once done with it

	// matcher will match on method, URL and body
	r.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		if r.Body == nil {
			return cassette.DefaultMatcher(r, i)
		}
		var b bytes.Buffer
		if _, err := b.ReadFrom(r.Body); err != nil {
			return false
		}
		r.Body = io.NopCloser(&b)
		return cassette.DefaultMatcher(r, i) && (b.String() == "" || b.String() == i.Body)
	})

	httpClient := &http.Client{Transport: r}

	f(httpClient)
}
