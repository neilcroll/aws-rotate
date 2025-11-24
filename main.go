/*
 * Project Name: AWS Multi-Account Credentials Rotator
 * Description: This tool helps you seamlessly rotate your AWS credentials
 * across multiple AWS accounts, ensuring enhanced security and compliance.
 *
 * Copyright (c) 2024 Neil Croll
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

type Profile struct {
	Name            string
	AccessKeyID     string
	SecretAccessKey string
}

const version = "v0.2.0"

// ANSI escape codes for bold text
const boldStart = "\033[1m"
const boldEnd = "\033[0m"

func main() {
	concurrent := flag.Bool("concurrent", true, "Run updates concurrently")
	sequential := flag.Bool("sequential", false, "Run updates sequentially")

	flag.BoolVar(concurrent, "c", true, "Run updates concurrently")
	flag.BoolVar(sequential, "s", false, "Run updates sequentially")

	flag.Parse()

	// If both flags are set, default to concurrent
	if *concurrent && *sequential {
		*sequential = false
	}

	displayWelcomeMessage()

	for {
		profiles := getAWSProfiles()
		selectedProfiles := selectProfiles(profiles)
		if len(selectedProfiles) == 0 {
			fmt.Println("No profiles selected. Please select at least one profile.")
			continue
		}
		username := getUsername()
		newPassword := getPassword()
		backupCredentialsFile()
		if *concurrent {
			updateProfilesSequential(selectedProfiles, username, newPassword)
		} else {
			updateProfilesConcurrent(selectedProfiles, username, newPassword)
		}
		fmt.Println("AWS credentials updated successfully.")
		break
	}
}

/**
 * Display a welcome message to the user.
 */
func displayWelcomeMessage() {
	fmt.Println("************************************************************")
	fmt.Println(boldStart + " Welcome to the AWS Multi-Account Credentials Rotator!" + boldEnd)
	fmt.Println(" This tool will help you rotate your AWS credentials.")
	fmt.Println("\n " + boldStart + "Note:" + boldEnd + " Each profile configured using the AWS CLI is\n expected to relate to a separate AWS account, but use\n the same username.\n ")
	fmt.Println(" Version:", boldStart+version+boldEnd)
	fmt.Println("************************************************************")
	fmt.Println()
}

/**
 * Get the AWS profiles from the credentials file.
 */
func getAWSProfiles() []Profile {
	homeDir, _ := os.UserHomeDir()
	credentialsFile := filepath.Join(homeDir, ".aws", "credentials")
	file, err := os.Open(credentialsFile)
	if err != nil {
		fmt.Println("Error opening credentials file:", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var profiles []Profile
	var currentProfile *Profile

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentProfile != nil {
				profiles = append(profiles, *currentProfile)
			}
			profileName := line[1 : len(line)-1]
			currentProfile = &Profile{Name: profileName}
		} else if strings.HasPrefix(line, "aws_access_key_id") && currentProfile != nil {
			currentProfile.AccessKeyID = strings.TrimSpace(strings.Split(line, "=")[1])
		} else if strings.HasPrefix(line, "aws_secret_access_key") && currentProfile != nil {
			currentProfile.SecretAccessKey = strings.TrimSpace(strings.Split(line, "=")[1])
		}
	}
	if currentProfile != nil {
		profiles = append(profiles, *currentProfile)
	}
	return profiles
}

/**
 * Present a menu for the user to select the profiles to update.
 */
func selectProfiles(profiles []Profile) []Profile {
	const all = "All"
	const allExceptDefault = "All except Default\n Available Profiles:" // The newline here is just a kludge to separate the two sections

	options := []string{all, allExceptDefault}
	profileNames := []string{}
	for _, profile := range profiles {
		profileNames = append(profileNames, profile.Name)
		options = append(options, profile.Name)
	}

	var selectedOptions []string
	for {
		prompt := &survey.MultiSelect{
			Message: "Select profiles to update:",
			Options: options,
		}
		survey.AskOne(prompt, &selectedOptions)

		if len(selectedOptions) == 0 {
			fmt.Println("No profiles selected. Please select at least one profile.")
			continue
		}

		if contains(selectedOptions, all) {
			return profiles
		}

		if contains(selectedOptions, allExceptDefault) {
			var selectedProfiles []Profile
			for _, profile := range profiles {
				if profile.Name != "default" {
					selectedProfiles = append(selectedProfiles, profile)
				}
			}
			return selectedProfiles
		}

		var selectedProfiles []Profile
		for _, option := range selectedOptions {
			for _, profile := range profiles {
				if profile.Name == option {
					selectedProfiles = append(selectedProfiles, profile)
				}
			}
		}
		return selectedProfiles
	}
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

/**
 * Get the username from the user.
 */
func getUsername() string {
	fmt.Println("Enter the username for the AWS accounts:")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	return username
}

/**
 * Get the new password from the user.
 */
func getPassword() string {
	fmt.Println("Enter the new password for the AWS accounts:")
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	return password
}

/**
 * Backup the credentials file.
 */
func backupCredentialsFile() {
	homeDir, _ := os.UserHomeDir()
	credentialsFile := filepath.Join(homeDir, ".aws", "credentials")
	backupFile := credentialsFile + "-" + time.Now().Format("20060102-150405") + ".bak"
	input, err := os.Open(credentialsFile)
	if err != nil {
		fmt.Println("Error opening credentials file:", err)
		os.Exit(1)
	}
	defer input.Close()

	output, err := os.Create(backupFile)
	if err != nil {
		fmt.Println("Error creating backup file:", err)
		os.Exit(1)
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	if err != nil {
		fmt.Println("Error copying to backup file:", err)
		os.Exit(1)
	}
}

/**
 * Update the profiles concurrently.
 */
func updateProfilesConcurrent(profiles []Profile, username, newPassword string) {
	var wg sync.WaitGroup
	for _, profile := range profiles {
		wg.Add(1)
		go func(profile Profile) {
			defer wg.Done()
			updateProfile(profile, username, newPassword)
		}(profile)
	}
	wg.Wait()
}

/**
 * Update the profiles sequentially.
 */
func updateProfilesSequential(profiles []Profile, username, newPassword string) {
	for _, profile := range profiles {
		updateProfile(profile, username, newPassword)
	}
}

/**
 * This is where the bulk of the updates happen. This function creates a new
 * access key, updates the credentials file with the new access key, and then
 * deletes the old access key. It also updates the password for the user.
 */
func updateProfile(profile Profile, username, newPassword string) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: profile.Name,
	})
	if err != nil {
		fmt.Println("Error creating session for profile", profile.Name)
		return
	}

	iamSvc := iam.New(sess)

	newKeys, err := createNewAccessKey(iamSvc, username)
	if err != nil {
		fmt.Println("Error creating new access key for profile", profile.Name)
		return
	}

	updateCredentialsFile(profile.Name, *newKeys.AccessKeyId, *newKeys.SecretAccessKey)

	// Wait for the new access key to become active - 15 seconds was an arbitrary choice for max wait time.
	fmt.Println("Waiting for new access key to become active ( for profile:", profile.Name, ")...")
	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)
		newSess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(*newKeys.AccessKeyId, *newKeys.SecretAccessKey, ""),
			Region:      aws.String("us-east-1"),
		})
		if err != nil {
			fmt.Println("Error creating new session for profile", profile.Name)
			continue
		}

		newIamSvc := iam.New(newSess)
		_, err = newIamSvc.GetUser(&iam.GetUserInput{})
		if err == nil {
			err = deleteOldAccessKey(iamSvc, username, profile.AccessKeyID)
			if err != nil {
				fmt.Println("Error deleting old access key for profile", profile.Name)
				break
			}

			err = updatePassword(newIamSvc, username, newPassword)
			if err != nil {
				fmt.Println("Error updating password for profile", profile.Name)
				fmt.Println(err)
				break
			}
			break
		}
	}
	fmt.Printf("Profile %s updated successfully.\n", profile.Name)
}

/**
 * Create a new access key for the user.
 */
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

/**
 * Update the credentials file with the new access key.
 */
func updateCredentialsFile(profileName, accessKeyID, secretAccessKey string) {
	homeDir, _ := os.UserHomeDir()
	credentialsFile := filepath.Join(homeDir, ".aws", "credentials")

	input, err := os.ReadFile(credentialsFile)
	if err != nil {
		fmt.Println("Error reading credentials file")
		return
	}

	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "["+profileName+"]") {
			lines[i+1] = "aws_access_key_id = " + accessKeyID
			lines[i+2] = "aws_secret_access_key = " + secretAccessKey
			break
		}
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(credentialsFile, []byte(output), 0644)
	if err != nil {
		fmt.Println("Error writing credentials file")
	}
}

/**
 * Delete the old access key for the user.
 */
func deleteOldAccessKey(svc *iam.IAM, username, accessKeyID string) error {
	trimmedAccessKeyID := strings.TrimSpace(accessKeyID)
	input := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(trimmedAccessKeyID),
		UserName:    aws.String(username),
	}
	_, err := svc.DeleteAccessKey(input)
	return err
}

/**
 * Update the password for the user.
 */
func updatePassword(svc *iam.IAM, username, newPassword string) error {
	input := &iam.UpdateLoginProfileInput{
		UserName: aws.String(username),
		Password: aws.String(newPassword),
	}
	_, err := svc.UpdateLoginProfile(input)
	return err
}
