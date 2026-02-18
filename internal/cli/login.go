package cli

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/idestis/pipe/internal/auth"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Pipe Hub",
	Args:  noArgs("pipe login"),
	RunE: func(cmd *cobra.Command, args []string) error {
		existing, err := auth.LoadCredentials()
		if err != nil {
			return fmt.Errorf("reading credentials: %w", err)
		}
		if existing != nil {
			log.Warn("already logged in", "username", existing.Username)
			fmt.Print("Re-authenticate? [y/N] ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if answer := scanner.Text(); answer != "y" && answer != "Y" {
				return nil
			}
		}

		client := auth.NewClient(apiURL)
		info := auth.CollectDeviceInfo()

		resp, err := client.InitiateDeviceAuth(&auth.DeviceAuthRequest{
			ClientName:     info.ClientName,
			ClientOS:       info.ClientOS,
			ClientArch:     info.ClientArch,
			ClientHostname: info.ClientHostname,
		})
		if err != nil {
			return fmt.Errorf("could not reach Pipe Hub at %s: %w", apiURL, err)
		}

		fmt.Println()
		fmt.Println("Attempting to open your default browser.")
		fmt.Println("If the browser does not open or you wish to use a different device to authorize this request, open the following URL:")
		fmt.Printf("\n  %s\n", resp.VerificationURIComplete)
		fmt.Printf("\nThen enter the code:\n\n  %s\n\n", resp.UserCode)

		if err := browser.OpenURL(resp.VerificationURIComplete); err != nil {
			log.Warn("could not open browser")
		}

		fmt.Println("Waiting for authorization...")

		status, err := auth.PollForAuthorization(client, resp.DeviceCode, resp.Interval, resp.ExpiresIn)
		if err != nil {
			return err
		}

		username := ""
		if status.Username != nil {
			username = *status.Username
		}
		apiKey := ""
		if status.APIKey != nil {
			apiKey = *status.APIKey
		}

		creds := &auth.Credentials{
			APIKey:       apiKey,
			Username:     username,
			APIBaseURL:   apiURL,
			AuthorizedAt: time.Now(),
		}
		if err := auth.SaveCredentials(creds); err != nil {
			return fmt.Errorf("saving credentials: %w", err)
		}

		fmt.Printf("Successfully logged in as %s\n", username)
		return nil
	},
}
