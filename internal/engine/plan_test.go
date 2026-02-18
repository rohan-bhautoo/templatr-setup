package engine

import (
	"testing"
)

func TestVersionSatisfies(t *testing.T) {
	tests := []struct {
		installed string
		required  string
		want      bool
		wantErr   bool
	}{
		// Basic semver ranges
		{"25.2.1", ">=20.0.0", true, false},
		{"18.17.0", ">=20.0.0", false, false},
		{"20.0.0", ">=20.0.0", true, false},
		{"3.12.0", ">=3.12.0", true, false},
		{"3.11.0", ">=3.12.0", false, false},

		// Exact version
		{"21.0.0", "21.0.0", true, false},
		{"21.1.0", "21.0.0", false, false},

		// Caret range
		{"20.5.0", "^20.0.0", true, false},
		{"21.0.0", "^20.0.0", false, false},

		// Tilde range
		{"20.0.5", "~20.0.0", true, false},
		{"20.1.0", "~20.0.0", false, false},

		// "latest" always passes
		{"1.0.0", "latest", true, false},
		{"999.0.0", "latest", true, false},

		// Version unknown
		{"installed (version unknown)", ">=20.0.0", false, true},

		// Invalid installed version
		{"not-a-version", ">=20.0.0", false, true},

		// Invalid requirement
		{"20.0.0", "not-valid-constraint", false, true},

		// Greater than or equal with pre-release style versions
		{"1.26.0", ">=1.22.0", true, false},

		// Simple major version
		{"21.0.0", ">=21", true, false},
		{"20.0.0", ">=21", false, false},
	}

	for _, tt := range tests {
		got, err := versionSatisfies(tt.installed, tt.required)
		if (err != nil) != tt.wantErr {
			t.Errorf("versionSatisfies(%q, %q) error = %v, wantErr %v", tt.installed, tt.required, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("versionSatisfies(%q, %q) = %v, want %v", tt.installed, tt.required, got, tt.want)
		}
	}
}

func TestSetupPlan_NeedsAction(t *testing.T) {
	// Plan with all skipped
	plan := &SetupPlan{
		Runtimes: []RuntimePlan{
			{Action: ActionSkip},
			{Action: ActionSkip},
		},
	}
	if plan.NeedsAction() {
		t.Error("NeedsAction() should return false when all actions are skip")
	}

	// Plan with an install
	plan.Runtimes = append(plan.Runtimes, RuntimePlan{Action: ActionInstall})
	if !plan.NeedsAction() {
		t.Error("NeedsAction() should return true when there's an install action")
	}

	// Plan with only upgrade
	plan2 := &SetupPlan{
		Runtimes: []RuntimePlan{
			{Action: ActionUpgrade},
		},
	}
	if !plan2.NeedsAction() {
		t.Error("NeedsAction() should return true when there's an upgrade action")
	}

	// Empty plan
	plan3 := &SetupPlan{}
	if plan3.NeedsAction() {
		t.Error("NeedsAction() should return false for empty plan")
	}
}

func TestActionType_Icons(t *testing.T) {
	tests := []struct {
		action    ActionType
		wantIcon  string
		wantStatus string
	}{
		{ActionSkip, "OK", "[OK]"},
		{ActionInstall, "Install", "[MISSING]"},
		{ActionUpgrade, "Upgrade", "[UPGRADE]"},
	}

	for _, tt := range tests {
		if got := tt.action.ActionIcon(); got != tt.wantIcon {
			t.Errorf("ActionIcon(%q) = %q, want %q", tt.action, got, tt.wantIcon)
		}
		if got := tt.action.StatusIcon(); got != tt.wantStatus {
			t.Errorf("StatusIcon(%q) = %q, want %q", tt.action, got, tt.wantStatus)
		}
	}
}
