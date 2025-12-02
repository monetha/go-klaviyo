package klaviyo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"

	"github.com/monetha/go-klaviyo/internal/log"
	"github.com/monetha/go-klaviyo/models/event"
	"github.com/monetha/go-klaviyo/models/profile"
	"github.com/monetha/go-klaviyo/models/profile/updater"
	"github.com/monetha/go-klaviyo/operations/getprofiles"
)

const (
	restAPIHost              = "https://a.klaviyo.com/api"
	revision                 = "2025-04-15"
	profileType              = "profile"
	profilesPath             = "profiles"
	profileImportPath        = "profile-import"
	profileBulkImportJobType = "profile-bulk-import-job"
	profileBulkImportPath    = "profile-bulk-import-jobs"
	eventType                = "event"
	eventsPath               = "events"

	// Default retry configuration
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 60 * time.Second
	defaultRetryMax     = 4

	clientTimeout = 30 * time.Second
)

var (
	// ErrInvalidAPIKey indicates that the provided Klaviyo API key is either not specified or invalid.
	ErrInvalidAPIKey = errors.New("klaviyo: invalid or missing API key")

	// ErrTooManyRequests is returned by the client method when the endpoint is retried
	// the maximum number of times defined by defaultRetryMax and still fails.
	ErrTooManyRequests = errors.New("klaviyo: too many requests for calling endpoint")

	// ErrProfileDoesNotExist indicates that an attempt was made to retrieve a profile
	// that does not exist in Klaviyo.
	ErrProfileDoesNotExist = errors.New("klaviyo: a profile does not exist")
)

var (
	// Ensure that APIError implements the error interface.
	_ error = (*APIError)(nil)

	// Ensure that BadHTTPResponseError implements the error interface.
	_ error = (*BadHTTPResponseError)(nil)

	// Ensure that BadHTTPResponseError implements the Unwrap method for Go's errors.Is() and errors.As() functions.
	_ interface {
		Unwrap() error
	} = (*BadHTTPResponseError)(nil)

	// Ensure that BadHTTPResponseError implements the Cause method, typically used with pkg/errors.
	_ interface {
		Cause() error
	} = (*BadHTTPResponseError)(nil)
)

// APIError represents an error returned by the Klaviyo API.
type APIError struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Source struct {
		Pointer string `json:"pointer"`
	} `json:"source"`
	Meta struct {
		DuplicateProfileID string `json:"duplicate_profile_id,omitempty"`
	} `json:"meta,omitempty"`
}

// Error returns a human-readable representation of the APIError.
func (e *APIError) Error() string {
	return fmt.Sprintf("Klaviyo API Error (ID: %s, Status: %d, Code: %s) - %s: %s",
		e.Id, e.Status, e.Code, e.Title, e.Detail)
}

// ErrProfileAlreadyExists indicates that an attempt was made to create a profile
// that already exists in Klaviyo. It holds the ID of the duplicate profile.
type ErrProfileAlreadyExists struct {
	DuplicateProfileID string
}

// Error returns a string representation of the ErrProfileAlreadyExists error.
// It conforms to the error interface.
func (e *ErrProfileAlreadyExists) Error() string {
	return fmt.Sprintf("klaviyo: a profile already exists with one of these identifiers: %s", e.DuplicateProfileID)
}

// BadHTTPResponseError represents an error due to a bad HTTP response.
type BadHTTPResponseError struct {
	statusCode int
	body       []byte
	cause      error
}

// StatusCode returns the HTTP status code of the response.
func (e *BadHTTPResponseError) StatusCode() int { return e.statusCode }

// Body returns the body of the HTTP response.
func (e *BadHTTPResponseError) Body() []byte { return e.body }

// Error returns a human-readable representation of the BadHTTPResponseError.
func (e *BadHTTPResponseError) Error() string {
	return "klaviyo: bad HTTP response: " + e.cause.Error()
}

// Cause returns the underlying cause of the error.
func (e *BadHTTPResponseError) Cause() error { return e.cause }

// Unwrap provides compatibility for Go's errors.Is() and errors.As() functions.
func (e *BadHTTPResponseError) Unwrap() error { return e.cause }

// Client represents a Klaviyo client with methods to interact with the Klaviyo API.
type Client struct {
	APIKey     string
	httpClient *http.Client
	restAPIURL *url.URL
}

// New initializes a new Klaviyo client with the default http client.
func New(apiKey string, logger *zap.Logger) *Client {
	return NewWithClient(
		apiKey,
		logger,
		&http.Client{
			Timeout: clientTimeout,
		})
}

// NewWithClient initializes a new Klaviyo client with a custom http client.
func NewWithClient(apiKey string, logger *zap.Logger, httpClient *http.Client) *Client {
	retryableHTTPClient := &retryablehttp.Client{
		HTTPClient:   httpClient,
		Logger:       log.NewLeveledLogger(logger),
		RetryWaitMin: defaultRetryWaitMin,
		RetryWaitMax: defaultRetryWaitMax,
		RetryMax:     defaultRetryMax,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
		ErrorHandler: errorHandler,
	}

	restAPIURL, err := url.Parse(restAPIHost)
	if err != nil {
		panic(err)
	}

	return &Client{
		APIKey:     apiKey,
		httpClient: retryableHTTPClient.StandardClient(),
		restAPIURL: restAPIURL,
	}
}

// setCommonHeaders sets common headers required for Klaviyo API requests.
func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Klaviyo-API-Key "+c.APIKey)
	req.Header.Set("accept", "application/json")
	req.Header.Set("revision", revision)
}

// GetEvents retrieves a list of created events from Klaviyo.
func (c *Client) GetEvents(ctx context.Context, params ...getprofiles.Param) ([]*event.ExistingEvent, error) {
	fields := url.Values{}
	for _, p := range params {
		p.Apply(fields)
	}

	var result struct {
		Data []*event.ExistingEvent `json:"data"`
	}
	if err := c.doReq(ctx, http.MethodGet, eventsPath, fields, nil, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// CreateEvent creates a new event in Klaviyo.
func (c *Client) CreateEvent(ctx context.Context, e *event.NewEvent, ID string, metricName string) error {
	type requestData struct {
		*event.NewEvent
		Type string `json:"type"`
	}

	type reqProfile struct {
		*event.ExistingProfile
		Type string `json:"type"`
	}

	type reqMetric struct {
		Type string `json:"type"`
		*event.NewMetric
	}

	profileRequestData := struct {
		Data reqProfile `json:"data"`
	}{
		Data: reqProfile{
			Type:            profileType,
			ExistingProfile: &event.ExistingProfile{ID: ID},
		},
	}

	metricRequestData := struct {
		Data reqMetric `json:"data"`
	}{
		Data: reqMetric{
			Type: "metric",
			NewMetric: &event.NewMetric{
				Attributes: event.MetricAttributes{Name: metricName},
			},
		},
	}

	request := struct {
		Data requestData `json:"data"`
	}{
		Data: requestData{
			NewEvent: e,
			Type:     eventType,
		},
	}
	request.Data.NewAttributes.Profile = profileRequestData
	request.Data.NewAttributes.Metric = metricRequestData

	if err := c.doReq(ctx, http.MethodPost, eventsPath, nil, request, nil); err != nil {
		return err
	}

	return nil
}

// GetProfiles retrieves a list of created profiles from Klaviyo.
func (c *Client) GetProfiles(ctx context.Context, params ...getprofiles.Param) ([]*profile.ExistingProfile, error) {
	fields := url.Values{}
	for _, p := range params {
		p.Apply(fields)
	}

	var result struct {
		Data []*profile.ExistingProfile `json:"data"`
	}
	if err := c.doReq(ctx, http.MethodGet, profilesPath, fields, nil, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// CreateProfile creates a new profile in Klaviyo. If a profile with the same identifiers
// already exists, it will return ErrProfileAlreadyExists.
func (c *Client) CreateProfile(ctx context.Context, p *profile.NewProfile) (*profile.ExistingProfile, error) {
	type requestData struct {
		*profile.NewProfile
		Type string `json:"type"`
	}

	request := struct {
		Data requestData `json:"data"`
	}{
		Data: requestData{
			NewProfile: p,
			Type:       profileType,
		},
	}

	var result struct {
		Data profile.ExistingProfile `json:"data"`
	}
	if err := c.doReq(ctx, http.MethodPost, profilesPath, nil, request, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// GetProfile retrieves a specific profile by its ID from Klaviyo. If the profile
// with the given ID does not exist, it will return ErrProfileDoesNotExist.
func (c *Client) GetProfile(ctx context.Context, profileID string) (*profile.ExistingProfile, error) {
	endpoint := profilesPath + "/" + profileID + "/"

	var result struct {
		Data profile.ExistingProfile `json:"data"`
	}
	if err := c.doReq(ctx, http.MethodGet, endpoint, nil, nil, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// UpdateProfile updates a specific profile by its ID in Klaviyo.
func (c *Client) UpdateProfile(ctx context.Context, profileID string, updaters ...updater.Profile) (*profile.ExistingProfile, error) {
	// Create an empty profile data to hold the updates
	profileData := updater.NewProfileData()

	// Apply each updater to the profile map
	for _, u := range updaters {
		u.Apply(profileData)
	}

	// Create the request data structure
	type requestData struct {
		Attributes map[string]interface{} `json:"attributes"`
		Id         string                 `json:"id"`
		Type       string                 `json:"type"`
		Meta       map[string]interface{} `json:"meta,omitempty"`
	}

	var meta map[string]interface{}
	if propertiesToRemove := profileData.PropertiesToRemove; len(propertiesToRemove) > 0 {
		meta = map[string]interface{}{
			"patch_properties": map[string]interface{}{
				"unset": propertiesToRemove,
			},
		}
	}

	request := struct {
		Data requestData `json:"data"`
	}{
		Data: requestData{
			Attributes: profileData.Attributes,
			Id:         profileID,
			Type:       profileType,
			Meta:       meta,
		},
	}

	endpoint := path.Join(profilesPath, profileID)

	var result struct {
		Data profile.ExistingProfile `json:"data"`
	}
	if err := c.doReq(ctx, http.MethodPatch, endpoint, nil, request, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// CreateOrUpdateProfile creates a new profile or updates an existing one
// using Klaviyo’s POST /api/profile-import “upsert” endpoint.
//
// Behaviour:
//
//   - If profileID is an empty string the call is treated as a pure upsert:
//     Klaviyo will look for a profile matching any identifiers contained in
//     the supplied updaters and will update it, or create a new profile if
//     none is found.
//
//   - If profileID is non-empty the request explicitly targets that profile
//     while still allowing identifier updates (email, phone, external_id…).
//
// Only the attributes produced by the supplied updaters are sent; fields
// you don’t set are **omitted altogether**, guaranteeing we never clear
// data unintentionally (cf. “setting a field to null will clear it” in
// the API docs).
//
// The method mirrors UpdateProfile’s signature so that callers can write:
//
//	    _, err := c.CreateOrUpdateProfile(
//		    ctx,
//		    klaviyoProfileID,   // "" if unknown
//		    klaviyoProfile.ToUpdaters()...,
//	    )
func (c *Client) CreateOrUpdateProfile(
	ctx context.Context,
	profileID string,
	updaters ...updater.Profile,
) (*profile.ExistingProfile, error) {
	// Collect attribute / meta mutations from updaters
	pd := updater.NewProfileData()
	for _, u := range updaters {
		u.Apply(pd)
	}

	// Build request payload
	type requestData struct {
		Attributes map[string]interface{} `json:"attributes"`
		Type       string                 `json:"type"`
		ID         string                 `json:"id,omitempty"`
		Meta       map[string]interface{} `json:"meta,omitempty"`
	}
	var meta map[string]interface{}
	if unset := pd.PropertiesToRemove; len(unset) > 0 {
		meta = map[string]interface{}{
			"patch_properties": map[string]interface{}{
				"unset": unset,
			},
		}
	}
	req := struct {
		Data requestData `json:"data"`
	}{
		Data: requestData{
			Attributes: pd.Attributes,
			Type:       profileType,
			ID:         profileID, // omitted if ""
			Meta:       meta,
		},
	}

	// Execute POST /api/profile-import
	var resp struct {
		Data profile.ExistingProfile `json:"data"`
	}
	if err := c.doReq(ctx, http.MethodPost, profileImportPath, nil, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// BulkProfileAttributes represents a profile in a bulk import request
type BulkProfileAttributes struct {
	Email      string                 `json:"email,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	// Add other fields as needed
	PhoneNumber *string `json:"phone_number,omitempty"`
	ExternalId  *string `json:"external_id,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
}

// BulkProfileData represents a single profile in bulk import
type BulkProfileData struct {
	Type       string                `json:"type"`
	Attributes BulkProfileAttributes `json:"attributes"`
}

// BulkImportJobAttributes contains the profiles array for bulk import
type BulkImportJobAttributes struct {
	Profiles struct {
		Data []BulkProfileData `json:"data"`
	} `json:"profiles"`
}

// BulkImportJobResponse represents the response from bulk import API
type BulkImportJobResponse struct {
	Data struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			Status         string     `json:"status"`
			CreatedAt      time.Time  `json:"created_at"`
			TotalCount     int        `json:"total_count"`
			CompletedCount int        `json:"completed_count"`
			FailedCount    int        `json:"failed_count"`
			CompletedAt    *time.Time `json:"completed_at"`
			ExpiresAt      time.Time  `json:"expires_at"`
			StartedAt      *time.Time `json:"started_at"`
		} `json:"attributes"`
	} `json:"data"`
}

// BulkCreateOrUpdateProfiles creates or updates multiple profiles using Klaviyo's bulk import API.
// This method submits up to 10,000 profiles in a single async job.
//
// Parameters:
//   - ctx: context for the request
//   - profiles: slice of BulkProfileAttributes (max 10,000 profiles, max 5MB payload)
//
// Returns:
//   - job ID string for tracking the async job status
//   - error if the submission fails
//
// Example:
//
//	profiles := []klaviyo.BulkProfileAttributes{
//	    {
//	        Email: "user1@example.com",
//	        Properties: map[string]interface{}{
//	            "last_login_at": "2025-12-01T10:00:00Z",
//	        },
//	    },
//	    {
//	        Email: "user2@example.com",
//	        Properties: map[string]interface{}{
//	            "last_transaction_at": "2025-11-20T15:30:00Z",
//	        },
//	    },
//	}
//	jobID, err := client.BulkCreateOrUpdateProfiles(ctx, profiles)
func (c *Client) BulkCreateOrUpdateProfiles(
	ctx context.Context,
	profiles []BulkProfileAttributes,
) (string, error) {
	if len(profiles) == 0 {
		return "", errors.New("klaviyo: profiles slice cannot be empty")
	}
	if len(profiles) > 10000 {
		return "", errors.New("klaviyo: maximum 10,000 profiles per bulk import job")
	}

	// Build profile data array
	profileDataArray := make([]BulkProfileData, len(profiles))
	for i, p := range profiles {
		profileDataArray[i] = BulkProfileData{
			Type:       profileType,
			Attributes: p,
		}
	}

	request := struct {
		Data struct {
			Type       string                  `json:"type"`
			Attributes BulkImportJobAttributes `json:"attributes"`
		} `json:"data"`
	}{}

	request.Data.Type = profileBulkImportJobType
	request.Data.Attributes.Profiles.Data = profileDataArray

	var resp BulkImportJobResponse
	if err := c.doReq(ctx, http.MethodPost, profileBulkImportPath, nil, request, &resp); err != nil {
		return "", err
	}

	return resp.Data.ID, nil
}

func (c *Client) doReq(ctx context.Context, method, endpoint string, fields url.Values, bodyData, result interface{}) error {
	uri := *c.restAPIURL
	uri.Path = path.Join(uri.Path, endpoint)
	uri.RawQuery = fields.Encode()

	var bodyBuffer io.Reader

	if bodyData != nil {
		jsonData, err := json.Marshal(bodyData)
		if err != nil {
			return err
		}
		bodyBuffer = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri.String(), bodyBuffer)
	if err != nil {
		return err
	}

	c.setCommonHeaders(req)
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut {
		req.Header.Set("content-type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	defer func() {
		// Drain and close the body to let the Transport reuse the connection
		_, _ = io.Copy(io.Discard, resp.Body)
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if statusCode := resp.StatusCode; statusCode < 200 || statusCode >= 300 {
		var errs struct {
			Errors []*APIError `json:"errors"`
		}
		if jsErr := json.Unmarshal(body, &errs); jsErr != nil {
			return &BadHTTPResponseError{
				statusCode: statusCode,
				body:       body,
				cause:      jsErr,
			}
		}

		err := &multierror.Error{}
		for _, er := range errs.Errors {
			err = multierror.Append(err, er)
		}
		if len(err.Errors) == 0 {
			return &APIError{
				Status: statusCode,
				Title:  "Bad HTTP status",
				Detail: (string)(body),
			}
		}

		return wrapAPIError(err.Unwrap())
	}
	if result != nil {
		return json.Unmarshal(body, result)
	}
	return nil
}

func errorHandler(resp *http.Response, err error, _ int) (*http.Response, error) {
	if err != nil {
		return resp, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return resp, ErrTooManyRequests
	}

	return resp, err
}

func wrapAPIError(err error) error {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusConflict:
			if apiErr.Code == "duplicate_profile" {
				return &ErrProfileAlreadyExists{DuplicateProfileID: apiErr.Meta.DuplicateProfileID}
			}
		case http.StatusNotFound:
			if apiErr.Code == "not_found" {
				return ErrProfileDoesNotExist
			}
		case http.StatusUnauthorized:
			if apiErr.Code == "not_authenticated" || apiErr.Code == "authentication_failed" {
				return ErrInvalidAPIKey
			}
		}
	}
	return err
}
