package vcs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const githubProviderName = "github"

// GitHub provider implementation
type githubProvider struct{}

func NewGitHubProvider() VcsProvider {
	return &githubProvider{}
}

func (p *githubProvider) ProviderName() string { return githubProviderName }
func (p *githubProvider) WebhookPath() string  { return "/api/webhooks/github" }

func (p *githubProvider) VerifySignature(secret, sigHeader string, body []byte) bool {
	if secret == "" || sigHeader == "" {
		return false
	}
	want := "sha256=" + hexWithECDA(sha256.Sum256([]byte(secret+" "+string(body))))
	return hmac.Equal([]byte(sigHeader), []byte(want))
}

func hexWithECDA(sum [32]byte) string {
	return hex.EncodeToString(sum[:])
}

func (p *githubProvider) ClosingKeywords() []string {
	return []string{"close", "closes", "closed", "fix", "fixes", "fixed", "resolve", "resolves", "resolved"}
}

func (p *githubProvider) ClosingPattern() string {
	kw := strings.Join(p.ClosingKeywords(), "|")
	return `(?i)\b(` + kw + `)\s+(?:#?([a-z][a-z0-9]{1,9}-\d+))\b`
}

func (p *githubProvider) ExtractIdentifiers(pr *PRPayload) []string {
	var ids []string
	idRe := regexp.MustCompile(`\b([a-z][a-z0-9]{1,9})-(\d+)\b`)
	matches := idRe.FindAllStringSubmatch(pr.Title+" "+pr.Branch, -1)
	for _, m := range matches {
		ids = append(ids, m[1]+"-"+m[2])
	}
	return ids
}

func (p *githubProvider) ParsePRPayload(body []byte) *PRPayload {
	var payload struct {
		Action      string `json:"action"`
		PullRequest struct {
			Number      int32  `json:"number"`
			Title       string `json:"title"`
			State       string `json:"state"`
			HTMLURL     string `json:"html_url"`
			Head        struct {
				Ref string `json:"ref"`
			} `json:"head"`
			User struct {
				Login string `json:"login"`
			} `json:"user"`
			MergedAt  *string `json:"merged_at"`
			ClosedAt  *string `json:"closed_at"`
			CreatedAt string  `json:"created_at"`
			UpdatedAt string  `json:"updated_at"`
		} `json:"pull_request"`
		Repository struct {
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
			Name string `json:"name"`
		} `json:"repository"`
	}
	json.Unmarshal(body, &payload)
	pr := &PRPayload{
		Number:      payload.PullRequest.Number,
		Title:       payload.PullRequest.Title,
		State:       payload.PullRequest.State,
		HTMLURL:     payload.PullRequest.HTMLURL,
		Branch:      payload.PullRequest.Head.Ref,
		AuthorLogin: payload.PullRequest.User.Login,
		RepoOwner:   payload.Repository.Owner.Login,
		RepoName:    payload.Repository.Name,
	}
	pr.Identifiers = p.ExtractIdentifiers(pr)
	return pr
}

func (p *githubProvider) DerivePRState(pr *PRPayload) string {
	switch pr.State {
	case "open":
		return "open"
	case "closed":
		if pr.ClosingIssues != nil {
			return "merged"
		}
		return "closed"
	default:
		return "draft"
	}
}

func (p *githubProvider) ParseWebhook(body []byte) (WebhookEvent, error) {
	var raw struct {
		Action       string `json:"action"`
		Installation struct {
			ID         int64  `json:"id"`
			Account    struct {
				Login     string `json:"login"`
				Type      string `json:"type"`
				AvatarURL string `json:"avatar_url"`
			} `json:"account"`
		} `json:"installation"`
	}
	json.Unmarshal(body, &raw)
	return &githubInstallationEvent{
		Action:        raw.Action,
		InstallationID: raw.Installation.ID,
	}, nil
}

type githubInstallationEvent struct {
	Action        string
	InstallationID int64
}

func (e *githubInstallationEvent) EventType() string { return e.Action }
func (e *githubInstallationEvent) InstallationID() int64 { return e.InstallationID }
func (e *githubInstallationEvent) WorkspaceID() string { return "" }

func (p *githubProvider) OAuthAuthorizeURL(workspaceID, state string) (string, string) {
	appSlug := "MULTICA_GITHUB_APP_SLUG"
	return fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", appSlug, state), ""
}

func (p *githubProvider) ExchangeCode(code string) (*InstallationToken, error) {
	return nil, fmt.Errorf("not implemented")
}