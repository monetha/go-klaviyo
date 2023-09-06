# Go Klaviyo API Client

This project provides a Go client for interacting with the Klaviyo API. It's a partial implementation, tailored to specific features that were required. It's designed to be simple, efficient, and idiomatic.

## Features

- Profile Management: Fetch, create, update, and delete profiles.
- Built-in retry mechanisms for better reliability.
- Structured error handling.
- Easily extendable for additional endpoints.

## Installation

```bash
go get -u gitlab.com/monetha/go-klaviyo
```

## Usage

First, create a client:

```go
import "gitlab.com/monetha/go-klaviyo"

client := klaviyo.New(API_KEY, logger)
```

### Fetch Profiles

```go
profiles, err := client.GetProfiles(ctx)
```

### Create Profile

```go
newProfile := &profile.NewProfile{
    // populate your profile data
}
createdProfile, err := client.CreateProfile(ctx, newProfile)
```

### Fetch Profile by ID

```go
fetchedProfile, err := client.GetProfile(ctx, PROFILE_ID)
```

### Update Profile

```go
updates := []updater.Profile{
    // your updaters
}
updatedProfile, err := client.UpdateProfile(ctx, PROFILE_ID, updates...)
```

### Handling Errors

All errors returned by the client are structured. You can inspect the error to get more details:

```go
if errors.Is(err, klaviyo.ErrProfileAlreadyExists) {
    // Handle specific error
}
```

## Contributing
Contributions are welcome! Please feel free to submit a pull request, report an issue, or suggest additional features.

## License
This project is licensed under the MIT License - see the LICENSE file for details.