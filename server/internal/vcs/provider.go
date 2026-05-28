package vcs

import (
)

// WebhookEvent represents a parsed webhook event from a VCS provider.
type WebhookEvent interface {
	// EventType returns the type of event (e.g., "pull_request", "installation")
	EventType() string
	// InstallationID returns the installation ID if this is an installation event
	InstallationID() int64
	// WorkspaceID returns the workspace ID that owns this installation
	WorkspaceID() string
}

// InstallationToken represents the result of exchanging an OAuth code for an installation token.
type InstallationToken struct {
	AccessToken      string
	InstallationID   int64
	AccountLogin     string
	AccountType      string
	AccountAvatarURL string
}

// PRPayload represents a pull request from a webhook payload.
type PRPayload struct {
	Number      int32
	Title      string
	State      string
	HTMLURL    string
	Branch     string
	AuthorLogin string

	RepoOwner  string
	RepoName   string

	// For auto-linking
	Identifiers    []string
	ClosingIssues []string

	// CI status
	MergeableState string
	ChecksPassed  int64
	ChecksFailed  int64
}

// VcsProvider defines the interface for VCS platform integrations (GitHub, Gitee, etc.)
type VcsProvider interface {
	// ProviderName returns the unique name of this provider ("github", "gitee")
	ProviderName() string

	// WebhookPath returns the HTTP route path for this provider's webhook endpoint
	WebhookPath() string

	// VerifySignature validates the webhook signature from request headers
	VerifySignature(secret string, sigHeader string, body []byte) bool

	// ParseWebhook parses the webhook body and returns a WebhookEvent
	ParseWebhook(body []byte) (WebhookEvent, error)

	// OAuthAuthorizeURL returns the URL to redirect users to for OAuth authorization
	OAuthAuthorizeURL(workspaceID, state string) (string, error)

	// ExchangeCode exchanges an OAuth code for an installation token
	ExchangeCode(code string) (*InstallationToken, error)

	// ParsePRPayload extracts PR information from a webhook body
	ParsePRPayload(body []byte) *PRPayload

	// DerivePRState converts the raw PR state to multica's internal state
	DerivePRState(pr *PRPayload) string

	// ExtractIdentifiers extracts issue identifiers from PR title/body/branch
	ExtractIdentifiers(pr *PRPayload) []string

	// ClosingKeywords returns the keywords that indicate an issue should be closed
	ClosingKeywords() []string

	// ClosingPattern returns a regex pattern for matching closing keywords
	ClosingPattern() string
}

// Provider registry
var providers = make(map[string]VcsProvider)

// RegisterProvider registers a VCS provider
func RegisterProvider(p VcsProvider) {
	providers[p.ProviderName()] = p
}

// GetProvider returns a provider by name
func GetProvider(name string) VcsProvider {
	return providers[name]
}

// AllProviders returns all registered providers
func AllProviders() []VcsProvider {
	out := make([]VcsProvider, 0, len(providers))
	for _, p := range providers {
		out = append(out, p)
	}
	return out
}