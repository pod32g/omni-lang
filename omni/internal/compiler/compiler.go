package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/omni-lang/omni/internal/ast"
	cbackend "github.com/omni-lang/omni/internal/backend/c"
	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/mir/builder"
	"github.com/omni-lang/omni/internal/mir/printer"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/types/checker"
)

// Config captures the minimal inputs needed to drive the compilation pipeline.
type Config struct {
	InputPath  string
	OutputPath string
	Backend    string
	OptLevel   string
	Emit       string
	Dump       string
	DebugInfo  bool
}

// ErrNotImplemented indicates that a requested stage has not yet been implemented.
var ErrNotImplemented = errors.New("not implemented")

// Compile wires together the compiler pipeline. It currently serves as a thin
// placeholder until the real frontend, midend and backend are ready.
func Compile(cfg Config) error {
	if cfg.InputPath == "" {
		return fmt.Errorf("input path required")
	}

	backend := cfg.Backend
	if backend == "" {
		backend = "c"
	}
	if backend != "vm" && backend != "clift" && backend != "c" {
		return fmt.Errorf("unsupported backend: %s", backend)
	}

	emit := cfg.Emit
	if emit == "" {
		if backend == "vm" {
			emit = "mir"
		} else if backend == "c" {
			emit = "exe"
		} else {
			emit = "obj"
		}
	}

	switch backend {
	case "vm":
		if emit != "mir" {
			return fmt.Errorf("vm backend: emit option %q not supported", emit)
		}
	case "clift":
		if emit != "obj" && emit != "exe" && emit != "binary" && emit != "asm" {
			return fmt.Errorf("clift backend: emit option %q not supported", emit)
		}
	case "c":
		if emit != "exe" && emit != "asm" {
			return fmt.Errorf("c backend: emit option %q not supported", emit)
		}
	}

	if cfg.OutputPath != "" {
		if ext := filepath.Ext(cfg.OutputPath); ext == "" {
			// Allow executables without extensions
			if emit != "exe" && emit != "binary" {
				return fmt.Errorf("output path must include file extension")
			}
		}
	}

	src, err := os.ReadFile(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("read input %s: %w", cfg.InputPath, err)
	}

	mod, err := parser.Parse(cfg.InputPath, string(src))
	if err != nil {
		return err
	}

	// Merge locally imported modules' functions into the main module so the VM can resolve them
	if err := MergeImportedModules(mod, filepath.Dir(cfg.InputPath)); err != nil {
		return err
	}

	if err := checker.Check(cfg.InputPath, string(src), mod); err != nil {
		return err
	}

	mirMod, err := builder.BuildModule(mod)
	if err != nil {
		return err
	}

	// Run MIR passes (constant folding disabled temporarily due to loop variable issues)
	if err := passes.Verify(mirMod); err != nil {
		return err
	}
	// TODO: Re-enable constant folding with proper handling of mutable variables
	// pipeline := passes.NewPipeline("default")
	// if _, err := pipeline.Run(*mirMod); err != nil {
	// 	return err
	// }

	if cfg.Dump == "mir" {
		fmt.Println(printer.Format(mirMod))
	}

	switch backend {
	case "vm":
		return compileVM(cfg, emit, mirMod)
	case "clift":
		return compileCraneliftBackend(cfg, emit, mirMod)
	case "c":
		return compileCBackend(cfg, emit, mirMod)
	default:
		return fmt.Errorf("unsupported backend: %s", backend)
	}
}

// MergeImportedModules loads imported local modules and appends their function declarations
// into the root module with namespaced names (aliasOrSegment.funcName) so that calls like
// `math_utils.add` resolve at runtime. std imports are ignored here (handled as intrinsics).
func MergeImportedModules(mod *ast.Module, baseDir string) error {
	loader := NewModuleLoader()
	// Prefer absolute path of the current file directory first to avoid cwd issues
	if baseDir != "" {
		if abs, err := filepath.Abs(baseDir); err == nil {
			loader.AddSearchPath(abs)
		} else if baseDir != "." {
			loader.AddSearchPath(baseDir)
		}
	}

		// Add the omni std directory to search paths
		// Find the omni root directory by looking for the std directory
		if abs, err := filepath.Abs(baseDir); err == nil {
			// Walk up the directory tree to find the omni root
			current := abs
			for {
				stdPath := filepath.Join(current, "std")
				if _, err := os.Stat(stdPath); err == nil {
					// Check if this is the main std directory (contains std.omni)
					mainStdPath := filepath.Join(stdPath, "std.omni")
					if _, err := os.Stat(mainStdPath); err == nil {
						loader.AddSearchPath(current)
						break
					}
				}
				parent := filepath.Dir(current)
				if parent == current {
					break // Reached root
				}
				current = parent
			}
		}

	// Collect imports from both Module.Imports and top-level decls
	imports := make([]*ast.ImportDecl, 0, len(mod.Imports))
	imports = append(imports, mod.Imports...)
	for _, d := range mod.Decls {
		if imp, ok := d.(*ast.ImportDecl); ok {
			imports = append(imports, imp)
		}
	}

	for _, imp := range imports {
		if len(imp.Path) == 0 {
			continue
		}
		// Load all imports, including std imports
		imported, err := loader.LoadModule(imp.Path)
		if err != nil {
			return fmt.Errorf("load import %s: %w", strings.Join(imp.Path, "."), err)
		}

		// Process nested imports recursively for std modules
		if len(imp.Path) > 0 && imp.Path[0] == "std" {
			if err := mergeNestedImports(imported, loader, mod); err != nil {
				return fmt.Errorf("merge nested imports for %s: %w", strings.Join(imp.Path, "."), err)
			}
		}

		local := imp.Alias
		if local == "" {
			local = imp.Path[len(imp.Path)-1]
		}
		// Append cloned function decls with namespaced names
		for _, d := range imported.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok {
				cloned := *fn
				cloned.Name = local + "." + fn.Name
				mod.Decls = append(mod.Decls, &cloned)
			}
		}
	}
	return nil
}

// mergeNestedImports recursively processes nested imports in std modules
func mergeNestedImports(module *ast.Module, loader *ModuleLoader, targetMod *ast.Module) error {
	// Collect imports from both Module.Imports and top-level decls
	imports := make([]*ast.ImportDecl, 0, len(module.Imports))
	imports = append(imports, module.Imports...)
	for _, d := range module.Decls {
		if imp, ok := d.(*ast.ImportDecl); ok {
			imports = append(imports, imp)
		}
	}

	for _, imp := range imports {
		if len(imp.Path) == 0 {
			continue
		}
		// Only process std imports
		if len(imp.Path) > 0 && imp.Path[0] == "std" {
			// Load the nested std module
			imported, err := loader.LoadModule(imp.Path)
			if err != nil {
				return fmt.Errorf("load nested import %s: %w", strings.Join(imp.Path, "."), err)
			}

			// Recursively process its nested imports
			if err := mergeNestedImports(imported, loader, targetMod); err != nil {
				return fmt.Errorf("merge nested imports for %s: %w", strings.Join(imp.Path, "."), err)
			}

			// Merge the functions with the target module
			local := imp.Alias
			if local == "" {
				local = imp.Path[len(imp.Path)-1]
			}
			for _, d := range imported.Decls {
				if fn, ok := d.(*ast.FuncDecl); ok {
					cloned := *fn
					cloned.Name = local + "." + fn.Name
					targetMod.Decls = append(targetMod.Decls, &cloned)
				}
			}
		}
	}
	return nil
}

func compileVM(cfg Config, emit string, mod *mir.Module) error {
	output := cfg.OutputPath
	if output == "" {
		output = defaultOutputPath(cfg.InputPath, emit)
	}
	if err := ensureDir(output); err != nil {
		return err
	}

	rendered := printer.Format(mod)
	if !strings.HasSuffix(rendered, "\n") {
		rendered += "\n"
	}
	if err := os.WriteFile(output, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("write mir output: %w", err)
	}
	return nil
}

// compileCBackend compiles MIR using the C backend
func compileCBackend(cfg Config, emit string, mod *mir.Module) error {
	output := cfg.OutputPath
	if output == "" {
		output = defaultOutputPath(cfg.InputPath, emit)
	}
	if err := ensureDir(output); err != nil {
		return err
	}

	switch emit {
	case "exe":
		if cfg.DebugInfo {
			return compileCToExecutableWithDebug(mod, output, cfg.OptLevel, cfg.InputPath)
		} else if cfg.OptLevel != "O0" {
			return compileCToExecutableWithOpt(mod, output, cfg.OptLevel)
		} else {
			return compileCToExecutable(mod, output)
		}
	case "asm":
		return compileToAssembly(mod, output)
	default:
		return fmt.Errorf("c backend: emit option %q not supported", emit)
	}
}

// compileCToExecutable compiles MIR to executable using C backend
func compileCToExecutable(mod *mir.Module, outputPath string) error {
	// Generate C code
	cCode, err := cbackend.GenerateC(mod)
	if err != nil {
		return fmt.Errorf("failed to generate C code: %w", err)
	}

	// Write C code to temporary file
	cPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".c"
	if err := os.WriteFile(cPath, []byte(cCode), 0o644); err != nil {
		return fmt.Errorf("failed to write C code: %w", err)
	}

	// Compile C code to executable
	if err := compileCWrapper(cPath, outputPath); err != nil {
		return fmt.Errorf("failed to compile C code: %w", err)
	}

	// Clean up temporary file
	// os.Remove(cPath) // Temporarily commented out for debugging

	return nil
}

// compileCToExecutableWithOpt compiles MIR to optimized executable using C backend
func compileCToExecutableWithOpt(mod *mir.Module, outputPath string, optLevel string) error {
	// Generate optimized C code
	cCode, err := cbackend.GenerateCOptimized(mod, optLevel)
	if err != nil {
		return fmt.Errorf("failed to generate optimized C code: %w", err)
	}

	// Write C code to temporary file
	cPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".c"
	if err := os.WriteFile(cPath, []byte(cCode), 0o644); err != nil {
		return fmt.Errorf("failed to write C code: %w", err)
	}

	// Compile C code to executable with optimization
	if err := compileCWrapperWithOpt(cPath, outputPath, optLevel); err != nil {
		return fmt.Errorf("failed to compile optimized C code: %w", err)
	}

	// Clean up temporary file
	// os.Remove(cPath) // Temporarily commented out for debugging

	return nil
}

// compileCToExecutableWithDebug compiles MIR to debug executable using C backend
func compileCToExecutableWithDebug(mod *mir.Module, outputPath string, optLevel string, sourceFile string) error {
	// Generate C code with debug information
	gen := cbackend.NewCGeneratorWithDebug(mod, optLevel, true, sourceFile)
	cCode, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate C code with debug: %w", err)
	}

	// Write C code to temporary file
	cPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".c"
	if err := os.WriteFile(cPath, []byte(cCode), 0o644); err != nil {
		return fmt.Errorf("failed to write C code: %w", err)
	}

	// Generate source map if debug info is enabled
	sourceMap := gen.GenerateSourceMap()
	if sourceMap != nil {
		sourceMapPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".map"
		sourceMapJSON, err := json.MarshalIndent(sourceMap, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal source map: %w", err)
		}
		if err := os.WriteFile(sourceMapPath, sourceMapJSON, 0o644); err != nil {
			return fmt.Errorf("failed to write source map: %w", err)
		}
	}

	// Compile C code to executable with debug symbols
	if err := compileCWrapperWithDebug(cPath, outputPath, optLevel); err != nil {
		return fmt.Errorf("failed to compile C code with debug: %w", err)
	}

	// Clean up temporary file
	os.Remove(cPath)

	return nil
}

// compileCraneliftBackend compiles MIR to native code using Cranelift backend
func compileCraneliftBackend(cfg Config, emit string, mod *mir.Module) error {
	output := cfg.OutputPath
	if output == "" {
		output = defaultOutputPath(cfg.InputPath, emit)
	}
	if err := ensureDir(output); err != nil {
		return err
	}

	switch emit {
	case "obj":
		return compileToObject(mod, output)
	case "exe", "binary":
		if cfg.DebugInfo {
			return compileToExecutableWithDebug(mod, output, cfg.OptLevel, cfg.InputPath)
		}
		if cfg.OptLevel != "" {
			return compileToExecutableWithOpt(mod, output, cfg.OptLevel)
		}
		return compileToExecutable(mod, output)
	case "asm":
		return compileToAssembly(mod, output)
	default:
		return fmt.Errorf("unsupported emit format: %s", emit)
	}
}

func compileToObject(mod *mir.Module, outputPath string) error {
	// Convert MIR module to JSON
	jsonData, err := mod.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert MIR to JSON: %w", err)
	}

	// For now, create a placeholder object file with the JSON content
	// TODO: Implement actual Cranelift compilation when the Rust library is available
	content := fmt.Sprintf("# OmniLang Object File Placeholder\n# MIR JSON:\n%s\n", string(jsonData))

	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write object output: %w", err)
	}

	return nil
}

func compileToExecutable(mod *mir.Module, outputPath string) error {
	// First compile to object file
	objPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".o"
	if err := compileToObject(mod, objPath); err != nil {
		return fmt.Errorf("failed to compile to object: %w", err)
	}

	// Create a C wrapper that links with the runtime
	cPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".c"
	if err := generateCWrapper(mod, cPath); err != nil {
		return fmt.Errorf("failed to generate C wrapper: %w", err)
	}

	// Compile the C wrapper with the runtime
	if err := compileCWrapper(cPath, outputPath); err != nil {
		return fmt.Errorf("failed to compile C wrapper: %w", err)
	}

	// Clean up temporary files
	os.Remove(objPath)
	os.Remove(cPath)

	return nil
}

func compileToExecutableWithOpt(mod *mir.Module, outputPath string, optLevel string) error {
	// First compile to object file
	objPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".o"
	if err := compileToObject(mod, objPath); err != nil {
		return fmt.Errorf("failed to compile to object: %w", err)
	}

	// Create an optimized C wrapper that links with the runtime
	cPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".c"
	if err := generateCOptimizedWrapper(mod, cPath, optLevel); err != nil {
		return fmt.Errorf("failed to generate optimized C wrapper: %w", err)
	}

	// Compile the C wrapper with the runtime
	if err := compileCWrapperWithOpt(cPath, outputPath, optLevel); err != nil {
		return fmt.Errorf("failed to compile C wrapper: %w", err)
	}

	// Clean up temporary files
	os.Remove(objPath)
	os.Remove(cPath)

	return nil
}

func compileToExecutableWithDebug(mod *mir.Module, outputPath string, optLevel string, sourceFile string) error {
	// First compile to object file
	objPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".o"
	if err := compileToObject(mod, objPath); err != nil {
		return fmt.Errorf("failed to compile to object: %w", err)
	}

	// Create a C wrapper with debug information that links with the runtime
	cPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".c"
	if err := generateCWrapperWithDebug(mod, cPath, optLevel, true, sourceFile); err != nil {
		return fmt.Errorf("failed to generate C wrapper with debug: %w", err)
	}

	// Compile the C wrapper with the runtime and debug symbols
	if err := compileCWrapperWithDebug(cPath, outputPath, optLevel); err != nil {
		return fmt.Errorf("failed to compile C wrapper: %w", err)
	}

	// Clean up temporary files
	os.Remove(objPath)
	os.Remove(cPath)

	return nil
}

func compileToAssembly(mod *mir.Module, outputPath string) error {
	// First generate C code
	cPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".c"
	if err := generateCWrapper(mod, cPath); err != nil {
		return fmt.Errorf("failed to generate C code: %w", err)
	}

	// Compile C to assembly
	if err := compileCToAssembly(cPath, outputPath); err != nil {
		return fmt.Errorf("failed to compile C to assembly: %w", err)
	}

	// Clean up temporary C file
	os.Remove(cPath)

	return nil
}

func compileCToAssembly(cPath, asmPath string) error {
	// Find the runtime directory
	runtimeDir := findRuntimeDir()
	if runtimeDir == "" {
		return fmt.Errorf("runtime directory not found")
	}

	// Determine target platform
	targetOS, targetArch := getTargetPlatform()

	// First compile the runtime to assembly
	runtimeAsmPath := strings.TrimSuffix(asmPath, filepath.Ext(asmPath)) + "_rt.s"
	runtimeArgs := []string{
		"-S", // Generate assembly
		"-o", runtimeAsmPath,
		filepath.Join(runtimeDir, "omni_rt.c"),
		"-I", runtimeDir,
		"-std=c99",
		"-Wall",
		"-Wextra",
	}

	// Add platform-specific flags
	if targetOS == "windows" {
		runtimeArgs = append(runtimeArgs, "-DWINDOWS")
	} else if targetOS == "darwin" {
		runtimeArgs = append(runtimeArgs, "-DDARWIN")
	} else if targetOS == "linux" {
		runtimeArgs = append(runtimeArgs, "-DLINUX")
	}

	// Add architecture-specific flags
	if targetArch == "amd64" || targetArch == "x86_64" {
		runtimeArgs = append(runtimeArgs, "-DARCH_X86_64")
	} else if targetArch == "arm64" || targetArch == "aarch64" {
		runtimeArgs = append(runtimeArgs, "-DARCH_ARM64")
	}

	cmd := exec.Command("gcc", runtimeArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("runtime assembly compilation failed: %w", err)
	}

	// Then compile the main C file to assembly
	mainArgs := []string{
		"-S", // Generate assembly
		"-o", asmPath,
		cPath,
		"-I", runtimeDir,
		"-std=c99",
		"-Wall",
		"-Wextra",
	}

	// Add platform-specific flags
	if targetOS == "windows" {
		mainArgs = append(mainArgs, "-DWINDOWS")
	} else if targetOS == "darwin" {
		mainArgs = append(mainArgs, "-DDARWIN")
	} else if targetOS == "linux" {
		mainArgs = append(mainArgs, "-DLINUX")
	}

	// Add architecture-specific flags
	if targetArch == "amd64" || targetArch == "x86_64" {
		mainArgs = append(mainArgs, "-DARCH_X86_64")
	} else if targetArch == "arm64" || targetArch == "aarch64" {
		mainArgs = append(mainArgs, "-DARCH_ARM64")
	}

	cmd = exec.Command("gcc", mainArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("main assembly compilation failed: %w", err)
	}

	// Clean up runtime assembly file
	os.Remove(runtimeAsmPath)

	return nil
}

func compileCWrapperWithOpt(cPath, outputPath string, optLevel string) error {
	// Find the runtime directory
	runtimeDir := findRuntimeDir()
	if runtimeDir == "" {
		return fmt.Errorf("runtime directory not found")
	}

	// Determine target platform
	targetOS, targetArch := getTargetPlatform()

	// Compile with platform-specific flags and optimization
	args := []string{
		"-o", outputPath,
		cPath,
		filepath.Join(runtimeDir, "omni_rt.c"),
		"-I", runtimeDir,
		"-std=c99",
		"-Wall",
		"-Wextra",
		"-lm",
	}

	// Add optimization flags
	switch optLevel {
	case "0", "O0", "none":
		args = append(args, "-O0")
	case "1", "O1", "basic":
		args = append(args, "-O1")
	case "2", "O2", "standard":
		args = append(args, "-O2")
	case "3", "O3", "aggressive":
		args = append(args, "-O3")
	case "s", "Os", "size":
		args = append(args, "-Os")
	default:
		args = append(args, "-O2")
	}

	// Add platform-specific flags
	if targetOS == "windows" {
		args = append(args, "-DWINDOWS")
	} else if targetOS == "darwin" {
		args = append(args, "-DDARWIN")
	} else if targetOS == "linux" {
		args = append(args, "-DLINUX")
	}

	// Add architecture-specific flags
	if targetArch == "amd64" || targetArch == "x86_64" {
		args = append(args, "-DARCH_X86_64")
	} else if targetArch == "arm64" || targetArch == "aarch64" {
		args = append(args, "-DARCH_ARM64")
	}

	cmd := exec.Command("gcc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("c compilation failed: %w", err)
	}

	return nil
}

func compileCWrapperWithDebug(cPath, outputPath string, optLevel string) error {
	// Find the runtime directory
	runtimeDir := findRuntimeDir()
	if runtimeDir == "" {
		return fmt.Errorf("runtime directory not found")
	}

	// Determine target platform
	targetOS, targetArch := getTargetPlatform()

	// Compile with platform-specific flags, optimization, and debug symbols
	args := []string{
		"-o", outputPath,
		cPath,
		filepath.Join(runtimeDir, "omni_rt.c"),
		"-I", runtimeDir,
		"-std=c99",
		"-Wall",
		"-Wextra",
		"-g", // Generate debug symbols
		"-lm",
	}

	// Add optimization flags
	switch optLevel {
	case "0", "O0", "none":
		args = append(args, "-O0")
	case "1", "O1", "basic":
		args = append(args, "-O1")
	case "2", "O2", "standard":
		args = append(args, "-O2")
	case "3", "O3", "aggressive":
		args = append(args, "-O3")
	case "s", "Os", "size":
		args = append(args, "-Os")
	default:
		args = append(args, "-O2")
	}

	// Add platform-specific flags
	if targetOS == "windows" {
		args = append(args, "-DWINDOWS")
	} else if targetOS == "darwin" {
		args = append(args, "-DDARWIN")
	} else if targetOS == "linux" {
		args = append(args, "-DLINUX")
	}

	// Add architecture-specific flags
	if targetArch == "amd64" || targetArch == "x86_64" {
		args = append(args, "-DARCH_X86_64")
	} else if targetArch == "arm64" || targetArch == "aarch64" {
		args = append(args, "-DARCH_ARM64")
	}

	cmd := exec.Command("gcc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("c compilation with debug failed: %w", err)
	}

	return nil
}

func generateCWrapper(mod *mir.Module, cPath string) error {
	// Generate C code from MIR module
	cCode, err := cbackend.GenerateC(mod)
	if err != nil {
		return fmt.Errorf("generate C code: %w", err)
	}

	if err := os.WriteFile(cPath, []byte(cCode), 0o644); err != nil {
		return fmt.Errorf("write C wrapper: %w", err)
	}

	return nil
}

func generateCOptimizedWrapper(mod *mir.Module, cPath string, optLevel string) error {
	// Generate optimized C code from MIR module
	cCode, err := cbackend.GenerateCOptimized(mod, optLevel)
	if err != nil {
		return fmt.Errorf("generate optimized C code: %w", err)
	}

	if err := os.WriteFile(cPath, []byte(cCode), 0o644); err != nil {
		return fmt.Errorf("write C wrapper: %w", err)
	}

	return nil
}

func generateCWrapperWithDebug(mod *mir.Module, cPath string, optLevel string, debugInfo bool, sourceFile string) error {
	// Generate C code with debug information from MIR module
	cCode, err := cbackend.GenerateCWithDebug(mod, optLevel, debugInfo, sourceFile)
	if err != nil {
		return fmt.Errorf("generate C code with debug: %w", err)
	}

	if err := os.WriteFile(cPath, []byte(cCode), 0o644); err != nil {
		return fmt.Errorf("write C wrapper: %w", err)
	}

	return nil
}

func compileCWrapper(cPath, outputPath string) error {
	// Find the runtime directory
	runtimeDir := findRuntimeDir()
	if runtimeDir == "" {
		return fmt.Errorf("runtime directory not found")
	}

	// Determine target platform
	targetOS, targetArch := getTargetPlatform()

	// Compile with platform-specific flags
	args := []string{
		"-o", outputPath,
		cPath,
		filepath.Join(runtimeDir, "omni_rt.c"),
		"-I", runtimeDir,
		"-std=c99",
		"-Wall",
		"-Wextra",
		"-lm",
	}

	// Add platform-specific flags
	if targetOS == "windows" {
		args = append(args, "-DWINDOWS")
	} else if targetOS == "darwin" {
		args = append(args, "-DDARWIN")
	} else if targetOS == "linux" {
		args = append(args, "-DLINUX")
	}

	// Add architecture-specific flags
	if targetArch == "amd64" || targetArch == "x86_64" {
		args = append(args, "-DARCH_X86_64")
	} else if targetArch == "arm64" || targetArch == "aarch64" {
		args = append(args, "-DARCH_ARM64")
	}

	cmd := exec.Command("gcc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("c compilation failed: %w", err)
	}

	return nil
}

func getTargetPlatform() (string, string) {
	// Get current platform information
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go platform names to standard names
	osMap := map[string]string{
		"windows": "windows",
		"darwin":  "darwin",
		"linux":   "linux",
		"freebsd": "freebsd",
		"openbsd": "openbsd",
		"netbsd":  "netbsd",
	}

	archMap := map[string]string{
		"amd64":   "x86_64",
		"arm64":   "aarch64",
		"386":     "x86",
		"arm":     "arm",
		"ppc64":   "ppc64",
		"ppc64le": "ppc64le",
		"s390x":   "s390x",
	}

	targetOS := osMap[os]
	if targetOS == "" {
		targetOS = os
	}

	targetArch := archMap[arch]
	if targetArch == "" {
		targetArch = arch
	}

	return targetOS, targetArch
}

func findRuntimeDir() string {
	// Get the directory where the binary is located
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to current working directory
		execPath = "."
	}
	execDir := filepath.Dir(execPath)

	// Check if we're running in a test environment (temporary directory)
	isTestEnv := strings.Contains(execDir, "/tmp/") || strings.Contains(execDir, "go-build")

	// Try to find the runtime directory
	possiblePaths := []string{
		// First try relative to current working directory (works for tests)
		"./runtime",
		"../runtime",
		"../../runtime",
		"../../../runtime",
		"../../../../runtime",
	}

	// If not in test environment, also try relative to binary location
	if !isTestEnv {
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "runtime"),
			filepath.Join(execDir, "../runtime"),
			filepath.Join(execDir, "../../runtime"),
			filepath.Join(execDir, "../../../runtime"),
			filepath.Join(execDir, "../../../../runtime"),
		)
	}

	// Try to find the runtime directory
	for _, path := range possiblePaths {
		if _, err := os.Stat(filepath.Join(path, "omni_rt.h")); err == nil {
			return path
		}
	}

	// If still not found, try to find it by looking for the omni directory
	// This helps when running from different locations
	cwd, err := os.Getwd()
	if err == nil {
		// Walk up the directory tree looking for the omni directory
		current := cwd
		for i := 0; i < 10; i++ { // Limit to 10 levels up
			omniDir := filepath.Join(current, "omni")
			runtimePath := filepath.Join(omniDir, "runtime")
			if _, err := os.Stat(filepath.Join(runtimePath, "omni_rt.h")); err == nil {
				return runtimePath
			}
			parent := filepath.Dir(current)
			if parent == current {
				break // Reached root
			}
			current = parent
		}
	}

	return ""
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}

func defaultOutputPath(input, emit string) string {
	base := strings.TrimSuffix(input, filepath.Ext(input))
	switch emit {
	case "mir":
		return base + ".mir"
	case "asm":
		return base + ".s"
	case "obj":
		return base + ".o"
	case "exe", "binary":
		return base
	default:
		return base + ".out"
	}
}
