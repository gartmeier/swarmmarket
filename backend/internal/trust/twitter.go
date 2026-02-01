package trust

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// TwitterVerifier verifies tweets via the Twitter API.
type TwitterVerifier struct {
	bearerToken string
	httpClient  *http.Client
}

// NewTwitterVerifier creates a new Twitter verifier.
func NewTwitterVerifier(bearerToken string) *TwitterVerifier {
	return &TwitterVerifier{
		bearerToken: bearerToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifyTweet verifies that a tweet contains the expected challenge text.
// Returns the Twitter handle and user ID if successful.
func (t *TwitterVerifier) VerifyTweet(ctx context.Context, tweetURL, expectedText string) (twitterHandle, twitterUserID string, err error) {
	// Extract tweet ID from URL
	tweetID := extractTweetID(tweetURL)
	if tweetID == "" {
		return "", "", ErrInvalidTweetURL
	}

	// Fetch tweet via Twitter API v2
	url := fmt.Sprintf("https://api.twitter.com/2/tweets/%s?expansions=author_id&user.fields=username", tweetID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.bearerToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrTwitterAPIError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", "", ErrTweetNotFound
	}
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("%w: status %d", ErrTwitterAPIError, resp.StatusCode)
	}

	var result twitterAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Verify the tweet contains the expected text
	if !strings.Contains(result.Data.Text, expectedText) {
		return "", "", ErrTweetTextMismatch
	}

	// Get user info
	if len(result.Includes.Users) == 0 {
		return "", "", fmt.Errorf("could not get user info from tweet")
	}

	return result.Includes.Users[0].Username, result.Includes.Users[0].ID, nil
}

type twitterAPIResponse struct {
	Data struct {
		Text     string `json:"text"`
		AuthorID string `json:"author_id"`
	} `json:"data"`
	Includes struct {
		Users []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"users"`
	} `json:"includes"`
}

// extractTweetID extracts the tweet ID from a Twitter/X URL.
// Supports formats:
//   - https://twitter.com/user/status/1234567890
//   - https://x.com/user/status/1234567890
//   - https://mobile.twitter.com/user/status/1234567890
func extractTweetID(url string) string {
	re := regexp.MustCompile(`(?:twitter\.com|x\.com)/\w+/status/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
