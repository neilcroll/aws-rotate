package desktop

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"aws-rotate/internal/rotator"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App exposes backend functionality to the Wails runtime.
type App struct {
	ctx     context.Context
	service *rotator.Service
}

// ProfileSummary is sent to the frontend to describe an available profile.
type ProfileSummary struct {
	Name    string `json:"name"`
	IsChina bool   `json:"isChina"`
}

// ProfileSelection captures the user-selected state for each profile.
type ProfileSelection struct {
	Name    string `json:"name"`
	IsChina bool   `json:"isChina"`
	Checked bool   `json:"checked"`
}

// RotationRequest contains the data required to run a rotation from the UI.
type RotationRequest struct {
	Profiles []ProfileSelection `json:"profiles"`
	Username string             `json:"username"`
	Password string             `json:"password"`
	Mode     string             `json:"mode"`
}

// NewApp constructs the desktop app wrapper.
func NewApp() *App {
	return &App{
		service: rotator.NewService(rotator.Config{}),
	}
}

// Startup satisfies the Wails lifecycle hook.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ListProfiles returns the available AWS profiles for display in the UI.
func (a *App) ListProfiles() ([]ProfileSummary, error) {
	profiles, err := a.service.LoadProfiles()
	if err != nil {
		return nil, err
	}

	summaries := make([]ProfileSummary, len(profiles))
	for i, profile := range profiles {
		summaries[i] = ProfileSummary{
			Name: profile.Name,
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return summaries, nil
}

// RotateProfiles executes the rotation workflow for the selected profiles.
func (a *App) RotateProfiles(req RotationRequest) ([]rotator.ProfileResult, error) {
	if len(req.Profiles) == 0 {
		return nil, errors.New("no profiles were selected for rotation")
	}

	allProfiles, err := a.service.LoadProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS profiles from ~/.aws/credentials: %w", err)
	}

	profileMap := make(map[string]rotator.Profile, len(allProfiles))
	for _, profile := range allProfiles {
		profileMap[profile.Name] = profile
	}

	var selected []rotator.Profile
	for _, selection := range req.Profiles {
		if !selection.Checked {
			continue
		}
		profile, ok := profileMap[selection.Name]
		if !ok {
			return nil, fmt.Errorf("selected profile '%s' was not found in ~/.aws/credentials (file may have been modified)", selection.Name)
		}
		profile.IsChina = selection.IsChina
		selected = append(selected, profile)
	}

	if len(selected) == 0 {
		return nil, errors.New("no profiles were selected for rotation")
	}

	if _, err := a.service.BackupCredentialsFile(); err != nil {
		return nil, fmt.Errorf("failed to create backup of ~/.aws/credentials: %w (check file permissions and disk space)", err)
	}

	mode := parseMode(req.Mode)

	results, err := a.service.RotateProfiles(selected, req.Username, req.Password, rotator.Options{
		Mode: mode,
	})
	if err != nil {
		// Log catastrophic errors
		runtime.LogError(a.ctx, err.Error())
		return nil, err
	}

	// Results contain per-profile success/failure status
	// Log summary for debugging
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}
	runtime.LogInfo(a.ctx, fmt.Sprintf("Rotation complete: %d/%d profiles succeeded", successCount, len(results)))

	return results, nil
}

func parseMode(mode string) rotator.UpdateMode {
	switch rotator.UpdateMode(mode) {
	case rotator.ModeSequential:
		return rotator.ModeSequential
	default:
		return rotator.ModeConcurrent
	}
}
