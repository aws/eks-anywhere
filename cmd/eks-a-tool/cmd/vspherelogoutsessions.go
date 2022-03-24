package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var sessions []string

var vsphereSessionRmCommand = &cobra.Command{
	Use:    "sessions",
	Short:  "vsphere logout sessions command",
	Long:   "This command logs out all of the provided VSphere user sessions ",
	PreRun: prerunSessionLogoutCmd,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := vsphereLogoutSessions(cmd.Context(), sessions)
		if err != nil {
			log.Fatalf("Error removing sessions: %v", err)
		}
		return nil
	},
}

const (
	sessionTokensFlag      = "sessionTokens"
	tlsInsecureFlag        = "tlsInsecure"
	vsphereApiEndpointFlag = "vsphereApiEndpoint"
)

func init() {
	vsphereRmCmd.AddCommand(vsphereSessionRmCommand)
	vsphereSessionRmCommand.Flags().StringSliceVarP(&sessions, sessionTokensFlag, "s", []string{}, "sessions to logout")
	vsphereSessionRmCommand.Flags().Bool(tlsInsecureFlag, false, "if endpoint is tls secure or not")
	vsphereSessionRmCommand.Flags().StringP(vsphereApiEndpointFlag, "e", "", "the URL of the vsphere API endpoint")

	err := vsphereSessionRmCommand.MarkFlagRequired(vsphereApiEndpointFlag)
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}

	err = vsphereSessionRmCommand.MarkFlagRequired("sessionTokens")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func prerunSessionLogoutCmd(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func vsphereLogoutSessions(_ context.Context, sessions []string) error {
	failedSessionLogouts := map[string]error{}
	for _, session := range sessions {
		err := logoutSession(session)
		if err != nil {
			failedSessionLogouts[session] = err
		}
	}
	if len(failedSessionLogouts) > 0 {
		for k, v := range failedSessionLogouts {
			log.Printf("failed to log out session %s: %v", k, v)
		}
		return fmt.Errorf("failed to log %d sessions out of vsphere: %v", len(failedSessionLogouts), failedSessionLogouts)
	}
	return nil
}

func logoutSession(session string) error {
	log.Printf("logging out of session %s", session)
	sessionLogoutPayload := []byte(strings.TrimSpace(`
	<?xml version="1.0" encoding="UTF-8"?><Envelope xmlns="http://schemas.xmlsoap.org/soap/envelope/">
			<Body>
				<Logout xmlns="urn:vim25">
					<_this type="SessionManager">SessionManager</_this>
				</Logout>
			</Body>
	</Envelope>`,
	))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool(tlsInsecureFlag)},
	}
	client := &http.Client{Transport: tr}

	url := fmt.Sprintf("%s/sdk", viper.GetString(vsphereApiEndpointFlag))
	req, err := http.NewRequest("POST", url, bytes.NewReader(sessionLogoutPayload))
	if err != nil {
		return err
	}

	const sessionCookieKey = "vmware_soap_session"
	cookie := http.Cookie{Name: sessionCookieKey, Value: session}
	req.AddCookie(&cookie)

	req.Header.Set("Content-Type", "text/xml")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyString := string(bodyBytes)

	sessionNotAuthenticatedFault := "The session is not authenticated."
	if resp.StatusCode == 500 && strings.Contains(bodyString, sessionNotAuthenticatedFault) {
		log.Printf("Can't logout session %s, it's not logged in", session)
		return nil
	}
	if resp.StatusCode >= 499 {
		log.Printf("failed to log out of vsphere session %s: %v", session, bodyString)
		return fmt.Errorf("failed to log out of vsphere session %s: %v", session, bodyString)
	}
	log.Printf("Successfully logged out of vsphere session %s", session)
	return nil
}
