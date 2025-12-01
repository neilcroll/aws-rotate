package rotator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Profile represents one entry from the AWS credentials file.
type Profile struct {
	Name            string `json:"name"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	IsChina         bool   `json:"isChina"`
}

// Config configures the Service behaviour.
type Config struct {
	CredentialsPath string
}

// Service exposes helpers used by both CLI and desktop frontends.
type Service struct {
	credentialsPath string
}

// NewService builds a Service with sane defaults.
func NewService(cfg Config) *Service {
	path := cfg.CredentialsPath
	if path == "" {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, ".aws", "credentials")
	}
	return &Service{credentialsPath: path}
}

// LoadProfiles parses the credentials file into Profile structs.
func (s *Service) LoadProfiles() ([]Profile, error) {
	file, err := os.Open(s.credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("open credentials file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var (
		currentProfile *Profile
		profiles       []Profile
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]"):
			if currentProfile != nil {
				profiles = append(profiles, *currentProfile)
			}
			currentProfile = &Profile{Name: line[1 : len(line)-1]}
		case strings.HasPrefix(line, "aws_access_key_id") && currentProfile != nil:
			currentProfile.AccessKeyID = strings.TrimSpace(splitValue(line))
		case strings.HasPrefix(line, "aws_secret_access_key") && currentProfile != nil:
			currentProfile.SecretAccessKey = strings.TrimSpace(splitValue(line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan credentials: %w", err)
	}

	if currentProfile != nil {
		profiles = append(profiles, *currentProfile)
	}

	return profiles, nil
}

// UpdateCredentialsFile updates the access key pair for a profile in the credentials file.
func (s *Service) UpdateCredentialsFile(profileName, accessKeyID, secretAccessKey string) error {
	content, err := os.ReadFile(s.credentialsPath)
	if err != nil {
		return fmt.Errorf("read credentials: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	profileHeader := "[" + profileName + "]"
	updated := false

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == profileHeader {
			if i+2 >= len(lines) {
				return errors.New("credentials file missing key entries")
			}
			lines[i+1] = "aws_access_key_id = " + accessKeyID
			lines[i+2] = "aws_secret_access_key = " + secretAccessKey
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("profile %s not found", profileName)
	}

	output := strings.Join(lines, "\n")
	if err := os.WriteFile(s.credentialsPath, []byte(output), 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return nil
}

// BackupCredentialsFile takes a timestamped backup of the credentials file.
func (s *Service) BackupCredentialsFile() (string, error) {
	backupFile := s.credentialsPath + "-" + time.Now().Format("20060102-150405") + ".bak"

	input, err := os.Open(s.credentialsPath)
	if err != nil {
		return "", fmt.Errorf("open credentials: %w", err)
	}
	defer input.Close()

	output, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return "", fmt.Errorf("copy backup: %w", err)
	}

	return backupFile, nil
}

func splitValue(line string) string {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
