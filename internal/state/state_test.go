package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewState(t *testing.T) {
	s := NewState()
	if s.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", s.Version)
	}
	if len(s.Installations) != 0 {
		t.Errorf("expected 0 installations, got %d", len(s.Installations))
	}
	if len(s.PathModifications) != 0 {
		t.Errorf("expected 0 path mods, got %d", len(s.PathModifications))
	}
}

func TestState_SaveAndLoad(t *testing.T) {
	// Create a temp directory for state file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "state.json")

	s := NewState()
	s.AddInstallation(Installation{
		Runtime:  "node",
		Version:  "22.14.0",
		Path:     "/home/user/.templatr/runtimes/node/22.14.0",
		Template: "saas-landing",
	})
	s.AddPathModification(PathModification{
		Method: "shell_rc",
		File:   "/home/user/.bashrc",
		Value:  "/home/user/.templatr/runtimes/node/22.14.0/bin",
	})

	// Save to temp file
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %s", err)
	}
	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		t.Fatalf("write failed: %s", err)
	}

	// Load from temp file
	loadedData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read failed: %s", err)
	}

	var loaded State
	if err := json.Unmarshal(loadedData, &loaded); err != nil {
		t.Fatalf("unmarshal failed: %s", err)
	}

	if len(loaded.Installations) != 1 {
		t.Fatalf("expected 1 installation, got %d", len(loaded.Installations))
	}
	if loaded.Installations[0].Runtime != "node" {
		t.Errorf("expected runtime 'node', got %q", loaded.Installations[0].Runtime)
	}
	if loaded.Installations[0].Version != "22.14.0" {
		t.Errorf("expected version '22.14.0', got %q", loaded.Installations[0].Version)
	}
	if loaded.Installations[0].InstalledAt == "" {
		t.Error("expected InstalledAt to be set")
	}

	if len(loaded.PathModifications) != 1 {
		t.Fatalf("expected 1 path mod, got %d", len(loaded.PathModifications))
	}
	if loaded.PathModifications[0].Method != "shell_rc" {
		t.Errorf("expected method 'shell_rc', got %q", loaded.PathModifications[0].Method)
	}
}

func TestState_AddInstallation(t *testing.T) {
	s := NewState()
	s.AddInstallation(Installation{Runtime: "node", Version: "22.14.0"})
	s.AddInstallation(Installation{Runtime: "python", Version: "3.13.2"})

	if len(s.Installations) != 2 {
		t.Errorf("expected 2 installations, got %d", len(s.Installations))
	}

	// Both should have timestamps
	for _, inst := range s.Installations {
		if inst.InstalledAt == "" {
			t.Errorf("expected InstalledAt for %s", inst.Runtime)
		}
	}
}

func TestState_RemoveInstallation(t *testing.T) {
	s := NewState()
	s.AddInstallation(Installation{Runtime: "node", Version: "22.14.0"})
	s.AddInstallation(Installation{Runtime: "python", Version: "3.13.2"})

	s.RemoveInstallation("node", "22.14.0")

	if len(s.Installations) != 1 {
		t.Fatalf("expected 1 installation after removal, got %d", len(s.Installations))
	}
	if s.Installations[0].Runtime != "python" {
		t.Errorf("expected remaining installation to be python, got %s", s.Installations[0].Runtime)
	}
}

func TestState_GetInstallations(t *testing.T) {
	s := NewState()
	s.AddInstallation(Installation{Runtime: "node", Version: "22.14.0"})
	s.AddInstallation(Installation{Runtime: "node", Version: "20.11.0"})
	s.AddInstallation(Installation{Runtime: "python", Version: "3.13.2"})

	all := s.GetInstallations("")
	if len(all) != 3 {
		t.Errorf("expected 3 total, got %d", len(all))
	}

	nodeOnly := s.GetInstallations("node")
	if len(nodeOnly) != 2 {
		t.Errorf("expected 2 node installations, got %d", len(nodeOnly))
	}

	pythonOnly := s.GetInstallations("python")
	if len(pythonOnly) != 1 {
		t.Errorf("expected 1 python installation, got %d", len(pythonOnly))
	}

	goOnly := s.GetInstallations("go")
	if len(goOnly) != 0 {
		t.Errorf("expected 0 go installations, got %d", len(goOnly))
	}
}

func TestState_PathModifications(t *testing.T) {
	s := NewState()
	s.AddPathModification(PathModification{
		Method: "shell_rc",
		Value:  "/home/user/.templatr/runtimes/node/22.14.0/bin",
	})
	s.AddPathModification(PathModification{
		Method: "windows_env",
		Value:  `C:\Users\user\.templatr\runtimes\node\22.14.0`,
	})

	if len(s.PathModifications) != 2 {
		t.Fatalf("expected 2 path mods, got %d", len(s.PathModifications))
	}

	s.RemovePathModification("/home/user/.templatr/runtimes/node/22.14.0/bin")
	if len(s.PathModifications) != 1 {
		t.Fatalf("expected 1 path mod after removal, got %d", len(s.PathModifications))
	}
	if s.PathModifications[0].Method != "windows_env" {
		t.Errorf("expected remaining mod to be windows_env")
	}
}

func TestState_UndoInstallation(t *testing.T) {
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "node", "22.14.0")
	binDir := filepath.Join(installDir, "bin")

	// Create a fake installation directory
	os.MkdirAll(binDir, 0o755)
	os.WriteFile(filepath.Join(binDir, "node"), []byte("fake"), 0o755)

	s := NewState()
	s.AddInstallation(Installation{
		Runtime:         "node",
		Version:         "22.14.0",
		Path:            installDir,
		Action:          "upgrade",
		PreviousVersion: "20.11.0",
		PreviousPath:    "/usr/local/bin/node",
	})
	s.AddPathModification(PathModification{
		Method: "shell_rc",
		Value:  binDir,
	})

	result, err := s.UndoInstallation("node", "22.14.0")
	if err != nil {
		t.Fatalf("undo failed: %s", err)
	}

	// Directory should be removed
	if _, err := os.Stat(installDir); !os.IsNotExist(err) {
		t.Error("expected install directory to be removed")
	}

	// State should be cleaned up
	if len(s.Installations) != 0 {
		t.Errorf("expected 0 installations after undo, got %d", len(s.Installations))
	}

	// Should have PATH mod to clean up
	if result.PathMod == nil {
		t.Error("expected path mod in undo result")
	} else if result.PathMod.Value != binDir {
		t.Errorf("expected path mod value %q, got %q", binDir, result.PathMod.Value)
	}

	// Should have previous version info
	if result.Previous == nil {
		t.Error("expected previous version info")
	} else {
		if result.Previous.Version != "20.11.0" {
			t.Errorf("expected previous version 20.11.0, got %s", result.Previous.Version)
		}
	}
}

func TestState_UndoInstallation_NotFound(t *testing.T) {
	s := NewState()
	_, err := s.UndoInstallation("node", "22.14.0")
	if err == nil {
		t.Error("expected error for non-existent installation")
	}
}
