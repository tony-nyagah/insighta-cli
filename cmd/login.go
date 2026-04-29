package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"insighta-cli/internal/client"
	"insighta-cli/internal/credentials"
	"insighta-cli/internal/display"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via GitHub OAuth",
	RunE:  runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	apiURL := os.Getenv("INSIGHTA_API_URL")
	if apiURL == "" {
		apiURL = "https://api.insighta.app"
	}

	// 1. Generate PKCE values
	state, err := randomBase64(32)
	if err != nil {
		return err
	}
	codeVerifier, err := randomBase64(32)
	if err != nil {
		return err
	}
	codeChallenge := s256(codeVerifier)

	// 2. Find a free local port for the callback server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("could not start local callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// 3. Build the GitHub OAuth URL via the backend
	oauthURL := fmt.Sprintf(
		"%s/auth/github?state=%s&code_challenge=%s&code_challenge_method=S256&redirect_uri=%s",
		apiURL, state, codeChallenge, callbackURL,
	)

	// 4. Start local callback server
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	stateCh := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		receivedState := r.URL.Query().Get("state")
		code := r.URL.Query().Get("code")
		errParam := r.URL.Query().Get("error")

		if errParam != "" {
			fmt.Fprintf(w, "<html><body><h2>Authentication failed: %s</h2><p>You can close this tab.</p></body></html>", errParam)
			errCh <- fmt.Errorf("github oauth error: %s", errParam)
			return
		}

		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab and return to the terminal.</p></body></html>")
		stateCh <- receivedState
		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go func() {
		server.Serve(listener)
	}()

	// 5. Open browser
	display.Info(fmt.Sprintf("Opening GitHub in your browser...\n  %s\n", oauthURL))
	display.Info("Waiting for authentication (press Ctrl+C to cancel)")
	openBrowser(oauthURL)

	// 6. Wait for callback (2-minute timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	defer server.Shutdown(context.Background())

	var code string
	select {
	case <-ctx.Done():
		return fmt.Errorf("login timed out — please try again")
	case err := <-errCh:
		return err
	case receivedState := <-stateCh:
		if receivedState != state {
			return fmt.Errorf("state mismatch — possible CSRF attack, aborting")
		}
		code = <-codeCh
	}

	// 7. Exchange code + verifier with backend
	spin := display.NewSpinner("Exchanging code with backend...")
	spin.Start()

	exchangeURL := apiURL + "/auth/github/callback"
	raw, status, err := client.PostNoAuth(exchangeURL, map[string]string{
		"code":          code,
		"code_verifier": codeVerifier,
	})
	spin.Stop()

	if err != nil {
		return fmt.Errorf("exchange failed: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("authentication failed (HTTP %d): %s", status, string(raw))
	}

	// 8. Parse response and store credentials
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		} `json:"user"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("unexpected response: %w", err)
	}

	creds := &credentials.File{
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		Username:         result.User.Username,
		Role:             result.User.Role,
		AccessTokenExpAt: time.Now().Add(3 * time.Minute),
	}
	if err := credentials.Save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	display.Success(fmt.Sprintf("Logged in as @%s (%s)", result.User.Username, result.User.Role))
	return nil
}

// --- helpers ---

func randomBase64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func s256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		fmt.Printf("Open this URL in your browser:\n  %s\n", url)
		return
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
