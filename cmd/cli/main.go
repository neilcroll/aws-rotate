/*
 * Project Name: AWS Multi-Account Credentials Rotator
 * Description: CLI entrypoint that reuses the shared rotator service.
 */
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"aws-rotate/internal/rotator"

	"github.com/AlecAivazis/survey/v2"
)

const version = "v0.2.0"

// ANSI escape codes for bold text.
const (
	boldStart = "\033[1m"
	boldEnd   = "\033[0m"
)

func main() {
	concurrent := flag.Bool("concurrent", true, "Run updates concurrently")
	sequential := flag.Bool("sequential", false, "Run updates sequentially")

	flag.BoolVar(concurrent, "c", true, "Run updates concurrently")
	flag.BoolVar(sequential, "s", false, "Run updates sequentially")

	flag.Parse()

	mode := determineUpdateMode(*concurrent, *sequential)

	displayWelcomeMessage()

	service := rotator.NewService(rotator.Config{})

	for {
		profiles, err := service.LoadProfiles()
		if err != nil {
			fmt.Println("Failed to load AWS profiles:", err)
			os.Exit(1)
		}

		selected := selectProfiles(profiles)
		if len(selected) == 0 {
			fmt.Println("No profiles selected. Please select at least one profile.")
			continue
		}

		username := getUsername()
		newPassword := getPassword()

		if _, err := service.BackupCredentialsFile(); err != nil {
			fmt.Println("Failed to backup credentials file:", err)
			os.Exit(1)
		}

		results, err := service.RotateProfiles(selected, username, newPassword, rotator.Options{Mode: mode})
		for _, result := range results {
			if result.Success {
				fmt.Printf("✓ %s: %s\n", result.Name, result.Message)
			} else {
				fmt.Printf("✗ %s: %s\n", result.Name, result.Message)
			}
		}
		if err != nil {
			fmt.Println("One or more profiles failed to rotate.")
		} else {
			fmt.Println("AWS credentials updated successfully.")
		}
		break
	}
}

func determineUpdateMode(concurrent, sequential bool) rotator.UpdateMode {
	switch {
	case sequential && !concurrent:
		return rotator.ModeSequential
	default:
		return rotator.ModeConcurrent
	}
}

func displayWelcomeMessage() {
	fmt.Println("************************************************************")
	fmt.Println(boldStart + " Welcome to the AWS Multi-Account Credentials Rotator!" + boldEnd)
	fmt.Println(" This tool will help you rotate your AWS credentials.")
	fmt.Println("\n " + boldStart + "Note:" + boldEnd + " Each profile configured using the AWS CLI is\n expected to relate to a separate AWS account, but use\n the same username.\n ")
	fmt.Println(" Version:", boldStart+version+boldEnd)
	fmt.Println("************************************************************")
	fmt.Println()
}

func selectProfiles(profiles []rotator.Profile) []rotator.Profile {
	const all = "All"
	const allExceptDefault = "All except Default\n Available Profiles:"

	options := []string{all, allExceptDefault}
	for _, profile := range profiles {
		options = append(options, profile.Name)
	}

	var selectedOptions []string
	for {
		prompt := &survey.MultiSelect{
			Message: "Select profiles to update:",
			Options: options,
		}
		if err := survey.AskOne(prompt, &selectedOptions); err != nil {
			fmt.Println("Prompt failed:", err)
			return nil
		}

		if len(selectedOptions) == 0 {
			fmt.Println("No profiles selected. Please select at least one profile.")
			continue
		}

		if contains(selectedOptions, all) {
			return profiles
		}

		if contains(selectedOptions, allExceptDefault) {
			var selectedProfiles []rotator.Profile
			for _, profile := range profiles {
				if profile.Name != "default" {
					selectedProfiles = append(selectedProfiles, profile)
				}
			}
			return selectedProfiles
		}

		var selectedProfiles []rotator.Profile
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

func getUsername() string {
	fmt.Println("Enter the username for the AWS accounts:")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	return strings.TrimSpace(username)
}

func getPassword() string {
	fmt.Println("Enter the new password for the AWS accounts:")
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(password)
}
