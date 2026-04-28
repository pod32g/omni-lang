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
	"github.com/omni-lang/omni/internal/backend/cranelift"
	"github.com/omni-lang/omni/internal/logging"
	"github.com/omni-lang/omni/internal/mir"
	"github.com/omni-lang/omni/internal/mir/builder"
	"github.com/omni-lang/omni/internal/mir/printer"
	"github.com/omni-lang/omni/internal/parser"
	"github.com/omni-lang/omni/internal/passes"
	"github.com/omni-lang/omni/internal/types/checker"
)

// Config captures the minimal inputs needed to drive the compilation pipeline.
type Config struct {
	InputPath    string
	OutputPath   string
	Backend      string
	OptLevel     string
	Emit         string
	Dump         string
	DebugInfo    bool
	DebugModules bool
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
	if err := MergeImportedModules(mod, filepath.Dir(cfg.InputPath), cfg.DebugModules, backend); err != nil {
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
		logging.Logger().InfoFields("Using Cranelift backend", logging.String("emit", emit))
		return compileCraneliftBackend(cfg, emit, mirMod)
	case "c":
		return compileCBackend(cfg, emit, mirMod)
	default:
		return fmt.Errorf("unsupported backend: %s", backend)
	}
}

// MergeImportedModules loads imported local modules and appends their function declarations
// into the root module with namespaced names (aliasOrSegment.funcName) so that calls like
// `math_utils.add` resolve at runtime. std imports are ignored for C backend (handled as intrinsics)
// but loaded for VM backend.
func MergeImportedModules(mod *ast.Module, baseDir string, debugModules bool, backend string) error {
	loader := NewModuleLoader()
	logger := logging.Logger()

	// Add the base directory for local modules
	if baseDir != "" {
		if abs, err := filepath.Abs(baseDir); err == nil {
			loader.AddSearchPath(abs)
		} else if baseDir != "." {
			loader.AddSearchPath(baseDir)
		}
	}

	// Show debug information if requested
	if debugModules {
		logger.DebugString(loader.DebugInfo())
		logger.DebugString("Loading imports...")
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
		// Handle std imports based on backend
		if len(imp.Path) > 0 && imp.Path[0] == "std" {
			if backend == "vm" {
				// For VM backend, load std modules
				if debugModules {
					logger.DebugFields("Loading std import for VM", logging.String("path", strings.Join(imp.Path, ".")))
				}
				imported, err := loader.LoadModule(imp.Path)
				if err != nil {
					return fmt.Errorf("load std import %s: %w", strings.Join(imp.Path, "."), err)
				}

				aliases := make([]string, 0, 2)
				if imp.Alias != "" {
					aliases = append(aliases, imp.Alias)
				} else if len(imp.Path) > 0 {
					aliases = append(aliases, imp.Path[len(imp.Path)-1])
				}
				qualified := strings.Join(imp.Path, ".")
				additional := true
				for _, a := range aliases {
					if a == qualified {
						additional = false
						break
					}
				}
				if qualified != "" && additional {
					aliases = append(aliases, qualified)
				}

				// Recursively merge nested std imports used within this module
				if err := mergeNestedImports(imported, loader, mod, debugModules, map[string]bool{}); err != nil {
					return fmt.Errorf("merge nested imports for %s: %w", strings.Join(imp.Path, "."), err)
				}

				// Append cloned declarations with namespaced names for each alias
				for _, ns := range aliases {
					for _, d := range imported.Decls {
						switch decl := d.(type) {
						case *ast.FuncDecl:
							cloned := *decl
							cloned.Name = ns + "." + decl.Name
							mod.Decls = append(mod.Decls, &cloned)
						case *ast.StructDecl:
							cloned := *decl
							cloned.Name = ns + "." + decl.Name
							mod.Decls = append(mod.Decls, &cloned)
						case *ast.EnumDecl:
							cloned := *decl
							cloned.Name = ns + "." + decl.Name
							mod.Decls = append(mod.Decls, &cloned)
						case *ast.TypeAliasDecl:
							cloned := *decl
							cloned.Name = ns + "." + decl.Name
							mod.Decls = append(mod.Decls, &cloned)
						}
					}
				}
			} else {
				// For C backend, most std imports are handled as intrinsics
				// However, std.web contains many non-intrinsic functions that need to be compiled
				qualified := strings.Join(imp.Path, ".")
				// std submodules whose bodies (not just intrinsic signatures) need
				// to be compiled into the output. Pure-intrinsic modules like
				// std.math or std.io can stay on the "handled as intrinsic" path.
				// std submodules whose body-or-signature needs to be cloned
				// into the user module so the type checker can resolve
				// member access on the return values. For runtime-provided
				// functions (e.g. http_response_create) the C backend's
				// generateFunction skips emitting the body — only the
				// signature drives type info — but the cloning has to
				// happen first or member access falls back to void.
				//
				// std.network is included so the type checker sees
				// HTTPResponse / HTTPRequest return types on the
				// runtime-wired helpers (create/is_success/etc.).
				// The function bodies that *do* get cloned have to be
				// either pure-OmniLang (e.g. url_to_string) or marked
				// runtime-provided so the C backend skips their body.
				needsBodyLoad := qualified == "std.web" ||
					qualified == "std" ||
					qualified == "std.testing" ||
					qualified == "std.math" ||
					qualified == "std.collections" ||
					qualified == "std.network"
				if needsBodyLoad {
					// Load std or std.web module to include its functions (many are not intrinsics)
					if debugModules {
						logger.DebugFields("Loading std import for C backend", logging.String("path", strings.Join(imp.Path, ".")))
					}
					imported, err := loader.LoadModule(imp.Path)
					if err != nil {
						return fmt.Errorf("load std import %s: %w", strings.Join(imp.Path, "."), err)
					}

					// If importing std, also load std submodules whose bodies are needed.
					var webImported *ast.Module
					var extraSubmodules []struct {
						prefix string
						module *ast.Module
					}
					if qualified == "std" {
						webPath := []string{"std", "web"}
						var err error
						webImported, err = loader.LoadModule(webPath)
						if err != nil {
							webImported = nil
						}
						for _, name := range []string{"testing", "math", "collections", "network"} {
							sub, err := loader.LoadModule([]string{"std", name})
							if err == nil && sub != nil {
								extraSubmodules = append(extraSubmodules, struct {
									prefix string
									module *ast.Module
								}{prefix: "std." + name, module: sub})
							}
						}
					}

					aliases := make([]string, 0, 2)
					if imp.Alias != "" {
						aliases = append(aliases, imp.Alias)
					} else if len(imp.Path) > 0 {
						aliases = append(aliases, imp.Path[len(imp.Path)-1])
					}
					additional := true
					for _, a := range aliases {
						if a == qualified {
							additional = false
							break
						}
					}
					if qualified != "" && additional {
						aliases = append(aliases, qualified)
					}

					// For std.web, also add "web" alias
					if qualified == "std.web" {
						aliases = append(aliases, "web")
					}

					// Append cloned declarations with namespaced names for each alias
					for _, ns := range aliases {
						for _, d := range imported.Decls {
							switch decl := d.(type) {
							case *ast.FuncDecl:
								qname := ns + "." + decl.Name
								if isStdFunctionRuntimeProvided(qname) {
									continue
								}
								cloned := *decl
								cloned.Name = qname
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.StructDecl:
								cloned := *decl
								cloned.Name = ns + "." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.EnumDecl:
								cloned := *decl
								cloned.Name = ns + "." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.TypeAliasDecl:
								cloned := *decl
								cloned.Name = ns + "." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							}
						}
					}

					// If we loaded std.web, also add its declarations with std.web prefix
					if webImported != nil {
						// Collect web module types for qualification (both structs and type aliases)
						webTypes := make(map[string]bool)
						for _, d := range webImported.Decls {
							if structDecl, ok := d.(*ast.StructDecl); ok {
								webTypes[structDecl.Name] = true
							}
							if typeAlias, ok := d.(*ast.TypeAliasDecl); ok {
								webTypes[typeAlias.Name] = true
							}
						}
						
						for _, d := range webImported.Decls {
							switch decl := d.(type) {
							case *ast.FuncDecl:
								cloned := *decl
								cloned.Name = "std.web." + decl.Name
								// Update parameter types - qualify any web module types
								for i := range cloned.Params {
									if cloned.Params[i].Type != nil {
										typeName := cloned.Params[i].Type.Name
										// Check if this is a web module type (struct or type alias)
										if webTypes[typeName] {
											// Create a new TypeExpr with qualified name
											newType := *cloned.Params[i].Type
											newType.Name = "std.web." + typeName
											cloned.Params[i].Type = &newType
										}
									}
								}
								// Update return type - qualify any web module types
								if cloned.Return != nil {
									returnTypeName := cloned.Return.Name
									if webTypes[returnTypeName] {
										// Create a new TypeExpr with qualified name
										newReturn := *cloned.Return
										newReturn.Name = "std.web." + returnTypeName
										cloned.Return = &newReturn
									}
								}
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.StructDecl:
								cloned := *decl
								cloned.Name = "std.web." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.EnumDecl:
								cloned := *decl
								cloned.Name = "std.web." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.TypeAliasDecl:
								cloned := *decl
								cloned.Name = "std.web." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							}
						}
					}
					// Emit declarations for any additional submodules we loaded
					// transitively (std.testing, etc). Skip functions that are
					// already runtime-provided intrinsics to avoid duplicate
					// symbol errors at link time.
					for _, sub := range extraSubmodules {
						prefix := sub.prefix
						for _, d := range sub.module.Decls {
							switch decl := d.(type) {
							case *ast.FuncDecl:
								qname := prefix + "." + decl.Name
								if isStdFunctionRuntimeProvided(qname) {
									continue
								}
								cloned := *decl
								cloned.Name = qname
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.StructDecl:
								cloned := *decl
								cloned.Name = prefix + "." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.EnumDecl:
								cloned := *decl
								cloned.Name = prefix + "." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							case *ast.TypeAliasDecl:
								cloned := *decl
								cloned.Name = prefix + "." + decl.Name
								mod.Decls = append(mod.Decls, &cloned)
							}
						}
					}
				} else {
					// For other std imports, skip (handled as intrinsics)
					if debugModules {
						logger.DebugFields("Skipping std import (handled as intrinsic)", logging.String("path", strings.Join(imp.Path, ".")))
					}
				}
			}
			continue
		}
		// Load only local imports
		if debugModules {
			logger.DebugFields("Loading module", logging.String("path", strings.Join(imp.Path, ".")))
		}
		imported, err := loader.LoadModule(imp.Path)
		if err != nil {
			return fmt.Errorf("load import %s: %w", strings.Join(imp.Path, "."), err)
		}

		local := imp.Alias
		if local == "" {
			local = imp.Path[len(imp.Path)-1]
		}
		// Append cloned declarations with namespaced names (mirror std module behavior)
		for _, d := range imported.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				cloned := *decl
				cloned.Name = local + "." + decl.Name
				mod.Decls = append(mod.Decls, &cloned)
			case *ast.StructDecl:
				cloned := *decl
				cloned.Name = local + "." + decl.Name
				mod.Decls = append(mod.Decls, &cloned)
			case *ast.EnumDecl:
				cloned := *decl
				cloned.Name = local + "." + decl.Name
				mod.Decls = append(mod.Decls, &cloned)
			case *ast.TypeAliasDecl:
				cloned := *decl
				cloned.Name = local + "." + decl.Name
				mod.Decls = append(mod.Decls, &cloned)
			}
		}
	}
	return nil
}

// mergeNestedImports recursively processes nested imports in std modules
func mergeNestedImports(module *ast.Module, loader *ModuleLoader, targetMod *ast.Module, debugModules bool, visited map[string]bool) error {
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
			key := strings.Join(imp.Path, ".")
			if visited[key] {
				continue
			}
			visited[key] = true

			// Load the nested std module
			imported, err := loader.LoadModule(imp.Path)
			if err != nil {
				return fmt.Errorf("load nested import %s: %w", strings.Join(imp.Path, "."), err)
			}

			// Recursively process its nested imports
			if err := mergeNestedImports(imported, loader, targetMod, debugModules, visited); err != nil {
				return fmt.Errorf("merge nested imports for %s: %w", strings.Join(imp.Path, "."), err)
			}

			// Merge the functions with the target module
			aliases := make([]string, 0, 2)
			if imp.Alias != "" {
				aliases = append(aliases, imp.Alias)
			} else if len(imp.Path) > 0 {
				aliases = append(aliases, imp.Path[len(imp.Path)-1])
			}
			qualified := strings.Join(imp.Path, ".")
			additional := true
			for _, a := range aliases {
				if a == qualified {
					additional = false
					break
				}
			}
			if qualified != "" && additional {
				aliases = append(aliases, qualified)
			}

			for _, ns := range aliases {
				for _, d := range imported.Decls {
					switch decl := d.(type) {
					case *ast.FuncDecl:
						cloned := *decl
						cloned.Name = ns + "." + decl.Name
						targetMod.Decls = append(targetMod.Decls, &cloned)
					case *ast.StructDecl:
						cloned := *decl
						cloned.Name = ns + "." + decl.Name
						targetMod.Decls = append(targetMod.Decls, &cloned)
					case *ast.EnumDecl:
						cloned := *decl
						cloned.Name = ns + "." + decl.Name
						targetMod.Decls = append(targetMod.Decls, &cloned)
					case *ast.TypeAliasDecl:
						cloned := *decl
						cloned.Name = ns + "." + decl.Name
						targetMod.Decls = append(targetMod.Decls, &cloned)
					}
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
	os.Remove(cPath)

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
	os.Remove(cPath)

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
	// Use the Cranelift backend to compile MIR to object file
	return cranelift.CompileModuleToObject(mod, outputPath)
}

func compileToObjectWithOpt(mod *mir.Module, outputPath string, optLevel string) error {
	// Use the Cranelift backend to compile MIR to object file with optimization
	return cranelift.CompileModuleToObjectWithOpt(mod, outputPath, optLevel)
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
	// First compile to object file with optimization
	objPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".o"
	if err := compileToObjectWithOpt(mod, objPath, optLevel); err != nil {
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

	// Keep C file for debugging
	// os.Remove(cPath)

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
		"-pthread",
		// Silence codegen-emitted noise (unused SSA result slots, stub
		// bodies that don't read every parameter, the interface-dispatch
		// helper when no interface is used) while keeping the warnings
		// that catch real bugs — -Wuninitialized, -Wreturn-stack-address,
		// -Wreturn-type, etc. all stay on.
		"-Wno-unused-variable",
		"-Wno-unused-parameter",
		"-Wno-unused-function",
		// Every function gets an `entry:` label so self tail-calls can
		// `goto entry;` after reassigning params. Functions that don't
		// tail-recurse leave the label unused — that's expected, not a
		// bug.
		"-Wno-unused-label",
		// Encourage clang/gcc to apply sibling-call optimization to
		// cross-function tail calls. The C backend emits those as
		// `return f(args);`, which is the shape this optimization
		// recognizes. Self-recursion is already lowered to `goto entry;`
		// in codegen so it doesn't need the flag. No-op at -O0; most
		// useful at -O1 and above.
		"-foptimize-sibling-calls",
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
		"-pthread",
		// Silence codegen-emitted noise (unused SSA result slots, stub
		// bodies that don't read every parameter, the interface-dispatch
		// helper when no interface is used) while keeping the warnings
		// that catch real bugs — -Wuninitialized, -Wreturn-stack-address,
		// -Wreturn-type, etc. all stay on.
		"-Wno-unused-variable",
		"-Wno-unused-parameter",
		"-Wno-unused-function",
		// Every function gets an `entry:` label so self tail-calls can
		// `goto entry;` after reassigning params. Functions that don't
		// tail-recurse leave the label unused — that's expected, not a
		// bug.
		"-Wno-unused-label",
		// Encourage clang/gcc to apply sibling-call optimization to
		// cross-function tail calls. The C backend emits those as
		// `return f(args);`, which is the shape this optimization
		// recognizes. Self-recursion is already lowered to `goto entry;`
		// in codegen so it doesn't need the flag. No-op at -O0; most
		// useful at -O1 and above.
		"-foptimize-sibling-calls",
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
		"-pthread",
		// Silence codegen-emitted noise (unused SSA result slots, stub
		// bodies that don't read every parameter, the interface-dispatch
		// helper when no interface is used) while keeping the warnings
		// that catch real bugs — -Wuninitialized, -Wreturn-stack-address,
		// -Wreturn-type, etc. all stay on.
		"-Wno-unused-variable",
		"-Wno-unused-parameter",
		"-Wno-unused-function",
		// Every function gets an `entry:` label so self tail-calls can
		// `goto entry;` after reassigning params. Functions that don't
		// tail-recurse leave the label unused — that's expected, not a
		// bug.
		"-Wno-unused-label",
		// Encourage clang/gcc to apply sibling-call optimization to
		// cross-function tail calls. The C backend emits those as
		// `return f(args);`, which is the shape this optimization
		// recognizes. Self-recursion is already lowered to `goto entry;`
		// in codegen so it doesn't need the flag. No-op at -O0; most
		// useful at -O1 and above.
		"-foptimize-sibling-calls",
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

// isStdFunctionRuntimeProvided returns true when a qualified stdlib function
// name is implemented directly in the C runtime and must not be re-emitted from
// its Omni-level stub body (which would collide at link time).
func isStdFunctionRuntimeProvided(qualified string) bool {
	runtimeProvided := map[string]bool{
		// std.array — length is runtime-provided (omni_array_length);
		// get/set are handled at call sites with extra args.
		"std.array.length": true,
		// std.math — numeric intrinsics
		"std.math.abs":       true,
		"std.math.max":       true,
		"std.math.min":       true,
		"std.math.pow":       true,
		"std.math.sqrt":      true,
		"std.math.floor":     true,
		"std.math.ceil":      true,
		"std.math.round":     true,
		"std.math.trunc":     true,
		"std.math.cbrt":      true,
		"std.math.gcd":       true,
		"std.math.lcm":       true,
		"std.math.factorial": true,
		"std.math.sin":       true,
		"std.math.cos":       true,
		"std.math.tan":       true,
		"std.math.asin":      true,
		"std.math.acos":      true,
		"std.math.atan":      true,
		"std.math.atan2":     true,
		"std.math.log":       true,
		"std.math.log2":      true,
		"std.math.log10":     true,
		"std.math.exp":       true,
		"std.math.sinh":      true,
		"std.math.cosh":      true,
		"std.math.tanh":      true,
		"std.math.hypot":     true,
	}
	return runtimeProvided[qualified]
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
