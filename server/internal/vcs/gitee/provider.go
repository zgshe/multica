package vcs

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const giteeProviderName = "gitee"

// Gitee provider implementation
type giteeProvider struct{}

func NewGiteeProvider() VcsProvider {
	return &giteeProvider{}
}

func (p *giteeProvider) ProviderName() string { return giteeProviderName }
func (p *giteeProvider) WebhookPath() string { return "/api/webhooks/gitee" }

func (p *giteeProvider) VerifySignature(secret, tokenHeader string, body []byte) bool {
	if secret == "" || tokenHeader == "" {
		return false
	}
	// Gitee uses X-Gitee-Token which is a simple token comparison
	// or HMAC signature based on timestamp
	return hmac.Equal([]byte(tokenHeader), []byte(secret))
}

func (p *giteeProvider) ClosingKeywords() []string {
	return []string{"close", "closes", "closed", "fix", "fixes", "fixed", "resolve", "resolves", "resolved"}
}

func (p *giteeProvider) ClosingPattern() string {
	kw := strings.Join(p.ClosingKeywords(), "|")
	return `(?i)\b(` + kw + `)\s+(?:gitee\.com\/)?([a-zA-Z0-9_\-]+)\/(\d+)\b`
}

func (p *giteeProvider) ExtractIdentifiers(pr *PRPayload) []string {
	var ids []string
	// Gitee uses PROJECT-NUMBER format like GitHub
	idRe := regexp.MustCompile(`\b([a-z][a-z0-9]{1,9})-(\d+)\b`)
	matches := idRe.FindAllStringSubmatch(pr.Title+" "+pr.Branch, -1)
	for _, m := range matches {
		ids = append(ids, m[1]+"-"+m[2])
	}
	// Also support Gitee's numeric ID format: PROJECT/12345
	numRe := regexp.MustCompile(`([A-Za-z][A-Za-z0-9_\-]+)/(\d+)\b`)
	numMatches := numRe.FindAllStringSubmatch(pr.Title+" "+pr.Branch, -1)
	for _, m := range numMatches {
		ids = append(ids, m[1]+"-"+m[2])
	}
	return ids
}

func (p *giteeProvider) ParsePRPayload(body []byte) *PRPayload {
	var payload struct {
		Action     string `json:"action"`
		MergeRequest struct {
			Number    int32  `json:"number"`
			Title     string `json:"title"`
			State     string `json:"state"`
			HTMLURL   string `json:"html_url"`
			HeadRef   string `json:"head_ref"`
			Author    struct {
				Login string `json:"login"`
			} `json:"author"`
			MergedAt  *string `json:"merged_at"`
			ClosedAt  *string `json:"closed_at"`
			CreatedAt string  `json:"created_at"`
			UpdatedAt string  `json:"updated_at"`
		} `json:"merge_request"`
		Repository struct {
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
			Name string `json:"name"`
		} `json:"repository"`
	}
	json.Unmarshal(body, &payload)
	pr := &PRPayload{
		Number:      payload.MergeRequest.Number,
		Title:       payload.MergeRequest.Title,
		State:       payload.MergeRequest.State,
		HTMLURL:     payload.MergeRequest.HTMLURL,
		Branch:      payload.MergeRequest.HeadRef,
		AuthorLogin: payload.MergeRequest.Author.Login,
		RepoOwner:   payload.Repository.Owner.Login,
		RepoName:    payload.Repository.Name,
	}
	pr.Identifiers = p.ExtractIdentifiers(pr)
	return pr
}

func (p *giteeProvider) DerivePRState(pr *PRPayload) string {
	switch pr.State {
	case "open":
		return "open"
	case "merged":
		return "merged"
	case "closed":
		return "closed"
	default:
		return "draft"
	}
}

func (p *giteeProvider) ParseWebhook(body []byte) (WebhookEvent, error) {
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
	return &giteeInstallationEvent{
		Action:        raw.Action,
		InstallationID: raw.Installation.ID,
	}, nil
}

type giteeInstallationEvent struct {
	Action        string
	InstallationID int64
}

func (e *giteeInstallationEvent) EventType() string { return e.Action }
func (e *giteeInstallationEvent) InstallationID() int64 { return e.InstallationID }
func (e *giteeInstallationEvent) WorkspaceID() string { return "" }

func (p *giteeProvider) OAuthAuthorizeURL(workspaceID, state string) (string, error) {
	clientID := "MULTICA_GITEE_CLIENT_ID"
	return fmt.Sprintf("https://gitee.com/oauth/authorize?client_id=%s&response_type=code&redirect_uri=https://your-multica.com/api/gitee/setup&state=%s", clientID, state), nil
}

func (p *giteeProvider) ExchangeCode(code string) (*InstallationToken, error) {
	return nil, fmt.Errorf("not implemented")
}