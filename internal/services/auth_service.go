package services

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/localnerve/authorizer-go"
	"github.com/localnerve/propsdb/internal/config"
	"github.com/localnerve/propsdb/internal/utils"
)

var (
	authClient *authorizer.AuthorizerClient
	authOnce   sync.Once
)

// IsAuthorizerInitialized returns true if the Authorizer client is initialized
func IsAuthorizerInitialized() bool {
	return authClient != nil
}

// InitAuthorizer initializes the Authorizer client (singleton pattern)
func InitAuthorizer(cfg *config.Config, requestProtocol, requestHost string) error {
	var initErr error

	authOnce.Do(func() {
		// Ping the Authorizer service first
		if err := utils.PingAuthorizer(cfg.AuthzURL); err != nil {
			initErr = fmt.Errorf("authorizer ping failed: %w", err)
			return
		}

		redirectURL := fmt.Sprintf("%s://%s", requestProtocol, requestHost)
		log.Printf("Initializing Authorizer: authorizerURL=%s, clientID=%s, redirectURL=%s",
			cfg.AuthzURL, cfg.AuthzClientID, redirectURL)

		var err error
		authClient, err = authorizer.NewAuthorizerClient(cfg.AuthzClientID, cfg.AuthzURL, redirectURL, nil)
		if err != nil {
			initErr = fmt.Errorf("failed to create authorizer client: %w", err)
			return
		}
	})

	return initErr
}

// ValidateSession validates a session cookie or JWT for the given roles
func ValidateSession(token string, roles []string) (map[string]interface{}, error) {
	if authClient == nil {
		return nil, fmt.Errorf("authorizer client not initialized")
	}

	// Unescape the token just in case it was URI-encoded
	if unescaped, err := url.PathUnescape(token); err == nil {
		token = unescaped
	}

	// If the token contains spaces, it might be due to '+' being incorrectly decoded
	// Base64 tokens from Authorizer often contain '+'.
	if strings.Contains(token, " ") {
		token = strings.ReplaceAll(token, " ", "+")
	}

	// Convert roles to []*string
	rolesPtrs := make([]*string, len(roles))
	for i := range roles {
		rolesPtrs[i] = &roles[i]
	}

	// Check if the token is a JWT (contains dots)
	if strings.Contains(token, ".") {
		// Validate JWT using the authorizer-go SDK
		res, err := authClient.ValidateJWTToken(&authorizer.ValidateJWTTokenInput{
			Token:     token,
			TokenType: authorizer.TokenTypeAccessToken,
			Roles:     rolesPtrs,
		})

		if err != nil {
			return nil, fmt.Errorf("JWT validation failed: %w", err)
		}

		if res == nil || !res.IsValid {
			return nil, fmt.Errorf("JWT is not valid")
		}

		// To get the full User object, we call GetProfile with the token in the header
		user, err := authClient.GetProfile(map[string]string{
			"Authorization": "Bearer " + token,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user profile: %w", err)
		}

		return map[string]interface{}{
			"is_valid": true,
			"user":     user,
		}, nil
	}

	// Traditional session validation using the authorizer-go SDK
	res, err := authClient.ValidateSession(&authorizer.ValidateSessionInput{
		Cookie: token,
		Roles:  rolesPtrs,
	})

	if err != nil {
		return nil, fmt.Errorf("session validation failed: %w", err)
	}

	// Check if session is valid
	if res == nil || !res.IsValid {
		return nil, fmt.Errorf("session is not valid")
	}

	// Return user data
	return map[string]interface{}{
		"is_valid": res.IsValid,
		"user":     res.User,
	}, nil
}
