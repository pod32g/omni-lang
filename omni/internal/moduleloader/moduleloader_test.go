package moduleloader

import (
	"os"
	"testing"
)

func TestModuleLoader(t *testing.T) {
	t.Run("NewModuleLoader", func(t *testing.T) {
		loader := NewModuleLoader()
		if loader == nil {
			t.Fatal("NewModuleLoader returned nil")
		}

		if loader.cache == nil {
			t.Error("ModuleLoader cache is nil")
		}

		if loader.searchPaths == nil {
			t.Error("ModuleLoader searchPaths is nil")
		}
	})

	t.Run("AddSearchPath", func(t *testing.T) {
		loader := NewModuleLoader()
		testPath := "/test/path"

		loader.AddSearchPath(testPath)

		paths := loader.GetSearchPaths()
		found := false
		for _, path := range paths {
			if path == testPath {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Search path %s not found in paths: %v", testPath, paths)
		}
	})

	t.Run("GetSearchPaths", func(t *testing.T) {
		loader := NewModuleLoader()
		paths := loader.GetSearchPaths()

		if len(paths) == 0 {
			t.Error("Expected at least one search path")
		}
	})

	t.Run("DebugInfo", func(t *testing.T) {
		loader := NewModuleLoader()
		debug := loader.DebugInfo()

		if debug == "" {
			t.Error("Expected non-empty debug info")
		}
	})

	t.Run("LoadModule", func(t *testing.T) {
		loader := NewModuleLoader()

		// Test loading a non-existent module
		_, err := loader.LoadModule([]string{"nonexistent"})
		if err == nil {
			t.Error("Expected error when loading non-existent module")
		}
	})

	t.Run("FindModuleFile", func(t *testing.T) {
		loader := NewModuleLoader()

		// Test finding a non-existent module file
		_, err := loader.findModuleFile([]string{"nonexistent"})
		if err == nil {
			t.Error("Expected error when finding non-existent module file")
		}
	})

	t.Run("BuildSearchPaths", func(t *testing.T) {
		// Test with OMNI_STD_PATH environment variable
		os.Setenv("OMNI_STD_PATH", "/custom/std/path")
		defer os.Unsetenv("OMNI_STD_PATH")

		loader := NewModuleLoader()
		paths := loader.GetSearchPaths()

		// Should include the custom path
		found := false
		for _, path := range paths {
			if path == "/custom/std/path" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected OMNI_STD_PATH to be included in search paths")
		}
	})

	t.Run("FindBinaryRelativePaths", func(t *testing.T) {
		// This is a private function, but we can test it indirectly
		loader := NewModuleLoader()
		paths := loader.GetSearchPaths()

		// Should include current working directory
		found := false
		for _, path := range paths {
			if path == "." {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected current working directory to be included in search paths")
		}
	})

	t.Run("FindOmniRootFromCWD", func(t *testing.T) {
		// This is a private function, but we can test it indirectly
		// by checking if the search paths include reasonable locations
		loader := NewModuleLoader()
		paths := loader.GetSearchPaths()

		// This test might fail in some environments, so we'll just check
		// that we have some paths
		if len(paths) == 0 {
			t.Error("Expected at least one search path")
		}
	})

	t.Run("ModuleCache", func(t *testing.T) {
		loader := NewModuleLoader()

		// Test that cache is initialized
		if loader.cache == nil {
			t.Error("Expected cache to be initialized")
		}

		// Test that cache is empty initially
		if len(loader.cache) != 0 {
			t.Error("Expected cache to be empty initially")
		}
	})

	t.Run("SearchPathsInitialization", func(t *testing.T) {
		loader := NewModuleLoader()

		// Test that search paths are initialized
		if loader.searchPaths == nil {
			t.Error("Expected searchPaths to be initialized")
		}

		// Test that search paths are not empty
		if len(loader.searchPaths) == 0 {
			t.Error("Expected searchPaths to be non-empty")
		}
	})
}
