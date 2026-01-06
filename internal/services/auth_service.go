package services

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	//"io"
	"log"
	//"net/http"
	//"strings"
	"sync"
	//"time"

	"github.com/authorizerdev/authorizer-go"
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

	/*
		if err != nil && (strings.Contains(err.Error(), "selection of subfields") || strings.Contains(err.Error(), "GraphQL")) {
			// Fallback to raw GraphQL if SDK fails with selection error
			// We'll try to load config again since we don't have it here easily
			cfg, configErr := config.Load()
			if configErr == nil {
				return validateSessionRaw(cfg, cookie, roles)
			}
		}
	*/

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

/*
// validateSessionRaw performs a raw GraphQL request for session validation
func validateSessionRaw(cfg *config.Config, cookie string, roles []string) (map[string]interface{}, error) {
	// Simple query - some versions of authorizer read session from Cookie header
	query := `{
		validate_session {
			is_valid
			user {
				id
				email
			}
		}
	}`

	payload := map[string]interface{}{
		"query": query,
	}

	jsonPayload, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 5 * time.Second}
	// Append /graphql to base URL
	graphqlURL := strings.TrimSuffix(cfg.AuthzURL, "/") + "/graphql"
	req, err := http.NewRequest("POST", graphqlURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// Send the cookie in the Cookie header
	req.Header.Set("Cookie", "cookie_session="+cookie)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v, body: %s", err, string(body))
	}

	if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %v, body: %s", errors[0], string(body))
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no data in response, body: %s", string(body))
	}

	validateSession, ok := data["validate_session"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no validate_session in data, body: %s", string(body))
	}

	isValid, _ := validateSession["is_valid"].(bool)
	user, _ := validateSession["user"].(map[string]interface{})

	return map[string]interface{}{
		"is_valid": isValid,
		"user":     user,
	}, nil
}
*/
