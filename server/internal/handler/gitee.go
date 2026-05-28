package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// HandleGiteeWebhook handles incoming webhooks from Gitee
func (h *Handler) HandleGiteeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}

	secret := giteeWebhookSecret()
	if secret == "" {
		writeError(w, http.StatusServiceUnavailable, "gitee webhooks not configured")
		return
	}

	// Gitee uses X-Gitee-Token header for verification
	tokenHeader := r.Header.Get("X-Gitee-Token")
	if tokenHeader == "" {
		tokenHeader = r.Header.Get("X-Gitee-Signature")
	}
	if !verifyGiteeWebhookSignature(secret, tokenHeader, body) {
		writeError(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	event := r.Header.Get("X-Gitee-Event")
	ctx := r.Context()

	switch event {
	case "ping":
		writeJSON(w, http.StatusOK, map[string]string{"ok": "pong"})
		return
	case "Merge Request Hook":
		h.handleGiteeMergeRequestEvent(ctx, body)
	case "Note Pull Request Hook":
		// Comments on PRs - handle similarly to GitHub
		h.handleGiteeMergeRequestEvent(ctx, body)
	default:
		// Acknowledge other events
	}

	w.WriteHeader(http.StatusAccepted)
}

func verifyGiteeWebhookSignature(secret, header string, body []byte) bool {
	if secret == "" || header == "" {
		return false
	}
	// Gitee uses a token-based verification or HMAC-SHA256
	// For HMAC mode: signature = hex(hmac_sha256(secret, body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(header), []byte(expected))
}

func giteeWebhookSecret() string {
	// TODO: implement - should read from GITEE_WEBHOOK_SECRET env var
	return ""
}

func (h *Handler) handleGiteeMergeRequestEvent(ctx interface{}, body []byte) {
	var payload struct {
		Action string `json:"action"`
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
		Installation struct {
			ID int64 `json:"id"`
		} `json:"installation"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("failed to parse gitee merge request payload", "error", err)
		return
	}

	// Convert state from Gitee format
	state := payload.MergeRequest.State
	if state == "merged" || state == "merge" {
		state = "merged"
	} else if state == "closed" {
		state = "closed"
	} else {
		state = "open"
	}

	slog.Info("gitee merge request event",
		"action", payload.Action,
		"number", payload.MergeRequest.Number,
		"state", state,
		"repo", payload.Repository.Owner.Login+"/"+payload.Repository.Name,
	)

	// TODO: implement actual DB operations similar to GitHub handler
	// - upsert gitee_pull_request record
	// - auto-link to issues based on identifier regex
	// - handle close/merge to advance linked issues
}