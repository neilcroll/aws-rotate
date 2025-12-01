package rotator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aws-rotate/internal/rotator"
)

func TestLoadProfilesReadsAllProfiles(t *testing.T) {
	t.Parallel()

	creds := `
[default]
aws_access_key_id = DEFAULT123
aws_secret_access_key = secret-default

[dev]
aws_access_key_id = DEV123
aws_secret_access_key = secret-dev

[china-account]
aws_access_key_id = CHINA123
aws_secret_access_key = secret-china
`

	tempDir := t.TempDir()
	credPath := filepath.Join(tempDir, "credentials")
	if err := os.WriteFile(credPath, []byte(creds), 0o600); err != nil {
		t.Fatalf("write temp credentials: %v", err)
	}

	service := rotator.NewService(rotator.Config{CredentialsPath: credPath})
	profiles, err := service.LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v", err)
	}

	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}

	if profiles[1].Name != "dev" || profiles[1].AccessKeyID != "DEV123" {
		t.Fatalf("unexpected dev profile: %+v", profiles[1])
	}
}

func TestUpdateCredentialsFile(t *testing.T) {
	t.Parallel()

	creds := `
[dev]
aws_access_key_id = OLD
aws_secret_access_key = SECRET
`
	tempDir := t.TempDir()
	credPath := filepath.Join(tempDir, "credentials")
	if err := os.WriteFile(credPath, []byte(creds), 0o600); err != nil {
		t.Fatalf("write temp credentials: %v", err)
	}

	service := rotator.NewService(rotator.Config{CredentialsPath: credPath})
	if err := service.UpdateCredentialsFile("dev", "NEWKEY", "NEWSECRET"); err != nil {
		t.Fatalf("UpdateCredentialsFile() error = %v", err)
	}

	updated, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	got := string(updated)
	if !containsLine(got, "aws_access_key_id = NEWKEY") {
		t.Fatalf("access key not updated:\n%s", got)
	}
	if !containsLine(got, "aws_secret_access_key = NEWSECRET") {
		t.Fatalf("secret key not updated:\n%s", got)
	}
}

func containsLine(contents, line string) bool {
	return strings.Contains(contents, line)
}
