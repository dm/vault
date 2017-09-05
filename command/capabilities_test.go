package command

import (
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/cli"
)

func testCapabilitiesCommand(tb testing.TB) (*cli.MockUi, *CapabilitiesCommand) {
	tb.Helper()

	ui := cli.NewMockUi()
	return ui, &CapabilitiesCommand{
		BaseCommand: &BaseCommand{
			UI: ui,
		},
	}
}

func TestCapabilitiesCommand_Run(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		out  string
		code int
	}{
		{
			"too_many_args",
			[]string{"foo", "bar", "zip"},
			"Too many arguments",
			1,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ui, cmd := testCapabilitiesCommand(t)

			code := cmd.Run(tc.args)
			if code != tc.code {
				t.Errorf("expected %d to be %d", code, tc.code)
			}

			combined := ui.OutputWriter.String() + ui.ErrorWriter.String()
			if !strings.Contains(combined, tc.out) {
				t.Errorf("expected %q to contain %q", combined, tc.out)
			}
		})
	}

	t.Run("token", func(t *testing.T) {
		t.Parallel()

		client, closer := testVaultServer(t)
		defer closer()

		policy := `path "secret/foo" { capabilities = ["read"] }`
		if err := client.Sys().PutPolicy("policy", policy); err != nil {
			t.Error(err)
		}

		secret, err := client.Auth().Token().Create(&api.TokenCreateRequest{
			Policies: []string{"policy"},
			TTL:      "30m",
		})
		if err != nil {
			t.Fatal(err)
		}
		if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
			t.Fatalf("missing auth data: %#v", secret)
		}
		token := secret.Auth.ClientToken

		ui, cmd := testCapabilitiesCommand(t)
		cmd.client = client

		code := cmd.Run([]string{
			token, "secret/foo",
		})
		if exp := 0; code != exp {
			t.Errorf("expected %d to be %d", code, exp)
		}

		expected := "read"
		combined := ui.OutputWriter.String() + ui.ErrorWriter.String()
		if !strings.Contains(combined, expected) {
			t.Errorf("expected %q to contain %q", combined, expected)
		}
	})

	t.Run("local", func(t *testing.T) {
		t.Parallel()

		client, closer := testVaultServer(t)
		defer closer()

		policy := `path "secret/foo" { capabilities = ["read"] }`
		if err := client.Sys().PutPolicy("policy", policy); err != nil {
			t.Error(err)
		}

		secret, err := client.Auth().Token().Create(&api.TokenCreateRequest{
			Policies: []string{"policy"},
			TTL:      "30m",
		})
		if err != nil {
			t.Fatal(err)
		}
		if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
			t.Fatalf("missing auth data: %#v", secret)
		}
		token := secret.Auth.ClientToken

		client.SetToken(token)

		ui, cmd := testCapabilitiesCommand(t)
		cmd.client = client

		code := cmd.Run([]string{
			"secret/foo",
		})
		if exp := 0; code != exp {
			t.Errorf("expected %d to be %d", code, exp)
		}

		expected := "read"
		combined := ui.OutputWriter.String() + ui.ErrorWriter.String()
		if !strings.Contains(combined, expected) {
			t.Errorf("expected %q to contain %q", combined, expected)
		}
	})

	t.Run("communication_failure", func(t *testing.T) {
		t.Parallel()

		client, closer := testVaultServerBad(t)
		defer closer()

		ui, cmd := testCapabilitiesCommand(t)
		cmd.client = client

		code := cmd.Run([]string{
			"foo", "bar",
		})
		if exp := 2; code != exp {
			t.Errorf("expected %d to be %d", code, exp)
		}

		expected := "Error listing capabilities: "
		combined := ui.OutputWriter.String() + ui.ErrorWriter.String()
		if !strings.Contains(combined, expected) {
			t.Errorf("expected %q to contain %q", combined, expected)
		}
	})

	t.Run("no_tabs", func(t *testing.T) {
		t.Parallel()

		_, cmd := testCapabilitiesCommand(t)
		assertNoTabs(t, cmd)
	})
}
