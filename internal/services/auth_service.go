package services

import (
	"fmt"
	"log"
	"sync"

	"github.com/authorizerdev/authorizer-go"
	"github.com/localnerve/propsdb/internal/config"
	"github.com/localnerve/propsdb/internal/utils"
)

var (
	authClient *authorizer.AuthorizerClient
	authOnce   sync.Once
)

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

// ValidateSession validates a session cookie for the given roles
func ValidateSession(cookie string, roles []string) (map[string]interface{}, error) {
	if authClient == nil {
		return nil, fmt.Errorf("authorizer client not initialized")
	}

	// Convert roles to []*string
	rolesPtrs := make([]*string, len(roles))
	for i := range roles {
		rolesPtrs[i] = &roles[i]
	}

	// Validate session using the authorizer-go SDK
	res, err := authClient.ValidateSession(&authorizer.ValidateSessionInput{
		Cookie: cookie,
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
