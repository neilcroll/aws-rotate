package rotator

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// UpdateMode controls how profile updates are executed.
type UpdateMode string

const (
	// ModeConcurrent updates all selected profiles in parallel (default).
	ModeConcurrent UpdateMode = "concurrent"
	// ModeSequential updates profiles one at a time.
	ModeSequential UpdateMode = "sequential"

	defaultRegion  = "us-east-1"
	chinaRegion    = "cn-north-1"
	propagationMax = 15
)

// Options fine-tune the rotation behaviour.
type Options struct {
	Mode UpdateMode
}

// ProfileResult reports the outcome for a single profile rotation attempt.
type ProfileResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RotateProfiles runs the rotation workflow for the provided profiles.
func (s *Service) RotateProfiles(profiles []Profile, username, newPassword string, opts Options) ([]ProfileResult, error) {
	if len(profiles) == 0 {
		return nil, errors.New("no profiles provided")
	}
	if username == "" {
		return nil, errors.New("username is required")
	}
	if newPassword == "" {
		return nil, errors.New("new password is required")
	}
	if opts.Mode == "" {
		opts.Mode = ModeConcurrent
	}

	results := make([]ProfileResult, len(profiles))

	switch opts.Mode {
	case ModeSequential:
		for i, profile := range profiles {
			err := s.rotateProfile(profile, username, newPassword)
			results[i] = buildResult(profile.Name, err)
		}
	default:
		var wg sync.WaitGroup
		resultCh := make(chan ProfileResult, len(profiles))
		for _, profile := range profiles {
			wg.Add(1)
			go func(p Profile) {
				defer wg.Done()
				err := s.rotateProfile(p, username, newPassword)
				resultCh <- buildResult(p.Name, err)
			}(profile)
		}
		wg.Wait()
		close(resultCh)

		i := 0
		for result := range resultCh {
			results[i] = result
			i++
		}
	}

	var firstErr error
	for _, result := range results {
		if !result.Success {
			firstErr = errors.New("one or more profiles failed")
			break
		}
	}
	return results, firstErr
}

func (s *Service) rotateProfile(profile Profile, username, newPassword string) error {
	region := defaultRegion
	if profile.IsChina {
		region = chinaRegion
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: profile.Name,
		Config: aws.Config{
			Region: aws.String(region),
		},
	})
	if err != nil {
		return fmt.Errorf("profile %s: create session: %w", profile.Name, err)
	}

	iamSvc := iam.New(sess)

	newKeys, err := createNewAccessKey(iamSvc, username)
	if err != nil {
		return fmt.Errorf("profile %s: create new access key: %w", profile.Name, err)
	}

	if err := s.UpdateCredentialsFile(profile.Name, *newKeys.AccessKeyId, *newKeys.SecretAccessKey); err != nil {
		return fmt.Errorf("profile %s: update credentials file: %w", profile.Name, err)
	}

	newIamSvc, err := waitForPropagation(profile.Name, *newKeys.AccessKeyId, *newKeys.SecretAccessKey, region)
	if err != nil {
		return err
	}

	if err := deleteOldAccessKey(iamSvc, username, profile.AccessKeyID); err != nil {
		return fmt.Errorf("profile %s: delete old access key: %w", profile.Name, err)
	}

	if err := updatePassword(newIamSvc, username, newPassword); err != nil {
		return fmt.Errorf("profile %s: update password: %w", profile.Name, err)
	}

	return nil
}

func waitForPropagation(profileName, accessKeyID, secretKey, region string) (*iam.IAM, error) {
	for i := 0; i < propagationMax; i++ {
		time.Sleep(1 * time.Second)
		newSess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKeyID, secretKey, ""),
			Region:      aws.String(region),
		})
		if err != nil {
			continue
		}
		newIamSvc := iam.New(newSess)
		if _, err := newIamSvc.GetUser(&iam.GetUserInput{}); err == nil {
			return newIamSvc, nil
		}
	}
	return nil, fmt.Errorf("profile %s: new access key did not propagate in time", profileName)
}

func createNewAccessKey(svc *iam.IAM, username string) (*iam.AccessKey, error) {
	input := &iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	}
	result, err := svc.CreateAccessKey(input)
	if err != nil {
		return nil, err
	}
	return result.AccessKey, nil
}

func deleteOldAccessKey(svc *iam.IAM, username, accessKeyID string) error {
	trimmedAccessKeyID := strings.TrimSpace(accessKeyID)
	input := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(trimmedAccessKeyID),
		UserName:    aws.String(username),
	}
	_, err := svc.DeleteAccessKey(input)
	return err
}

func updatePassword(svc *iam.IAM, username, newPassword string) error {
	input := &iam.UpdateLoginProfileInput{
		UserName: aws.String(username),
		Password: aws.String(newPassword),
	}
	_, err := svc.UpdateLoginProfile(input)
	return err
}

func buildResult(name string, err error) ProfileResult {
	if err != nil {
		return ProfileResult{
			Name:    name,
			Success: false,
			Message: err.Error(),
		}
	}
	return ProfileResult{
		Name:    name,
		Success: true,
		Message: "Rotation completed",
	}
}
