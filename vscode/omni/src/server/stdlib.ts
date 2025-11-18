import * as fs from 'fs';
import * as path from 'path';
import { Connection, WorkspaceFolder } from 'vscode-languageserver/node';

export interface StdFunction {
  name: string;
  fullName: string; // e.g., "std.io.println"
  module: string; // e.g., "io"
  parameters: Array<{ name: string; type: string }>;
  returnType: string;
  documentation?: string;
  isAsync?: boolean;
}

export interface StdModule {
  name: string;
  functions: StdFunction[];
  submodules?: StdModule[];
}

export interface StdLibrary {
  modules: Map<string, StdModule>;
  functions: Map<string, StdFunction>;
}

// Find the standard library directory
function findStdDirectory(
  connection: Connection | null,
  workspaceFolders: WorkspaceFolder[] | null
): string | null {
  const possiblePaths: string[] = [];
  const triedPaths: string[] = [];

  // First, try paths relative to workspace root (most reliable)
  if (workspaceFolders && workspaceFolders.length > 0) {
    for (const folder of workspaceFolders) {
      // Convert URI to filesystem path
      let workspacePath = folder.uri;
      if (workspacePath.startsWith('file://')) {
        workspacePath = workspacePath.substring(7);
      }
      // Handle Windows paths
      if (process.platform === 'win32' && workspacePath.startsWith('/')) {
        workspacePath = workspacePath.substring(1);
      }
      
      const workspaceStdPath = path.join(workspacePath, 'omni/std');
      possiblePaths.push(workspaceStdPath);
    }
  }

  // Try environment variable path
  if (process.env.OMNI_STD_PATH) {
    possiblePaths.push(process.env.OMNI_STD_PATH);
  }

  // Try paths relative to current working directory
  possiblePaths.push(
    path.join(process.cwd(), 'omni/std'),
    path.join(process.cwd(), '../omni/std'),
    path.join(process.cwd(), '../../omni/std')
  );

  // Try paths relative to extension directory (when installed)
  // When compiled, __dirname will be in dist/server/
  possiblePaths.push(
    path.join(__dirname, '../../../../omni/std'),
    path.join(__dirname, '../../../omni/std'),
    path.join(__dirname, '../../../../../omni/std'),
    path.join(__dirname, '../../../../../../omni/std')
  );

  // Try paths relative to node_modules (if extension is installed)
  if (__dirname.includes('node_modules')) {
    const nodeModulesIndex = __dirname.indexOf('node_modules');
    const extensionRoot = path.join(__dirname.substring(0, nodeModulesIndex), 'omni/std');
    possiblePaths.push(extensionRoot);
  }

  // Try each path
  for (const stdPath of possiblePaths) {
    triedPaths.push(stdPath);
    try {
      const normalizedPath = path.normalize(stdPath);
      if (fs.existsSync(normalizedPath) && fs.statSync(normalizedPath).isDirectory()) {
        if (connection) {
          connection.console.log(`Found standard library at: ${normalizedPath}`);
        }
        return normalizedPath;
      }
    } catch (e) {
      // Ignore errors, continue trying
    }
  }

  // Log all tried paths if connection is available
  if (connection) {
    connection.console.warn('Standard library directory not found. Tried paths:');
    for (const triedPath of triedPaths) {
      connection.console.warn(`  - ${triedPath}`);
    }
  }

  return null;
}

// Parse function signature from OmniLang source
function parseFunctionSignature(
  text: string,
  moduleName: string,
  fullModuleName: string
): StdFunction | null {
  // Normalize whitespace - replace newlines and multiple spaces with single space
  const normalized = text.replace(/\s+/g, ' ').trim();
  
  // Match: async? func functionName(params):returnType
  // or: func functionName(params)
  // Try more flexible regex that handles multi-line signatures
  const funcRegex = /(async\s+)?func\s+(\w+)\s*\(([^)]*)\)\s*:?\s*([^{]*?)\s*\{/;
  let match = normalized.match(funcRegex);
  
  // If no match on normalized text, try original text (might have complex formatting)
  if (!match) {
    // Try to find function declaration across multiple lines
    const lines = text.split('\n');
    let funcLine = '';
    let inParams = false;
    let parenCount = 0;
    
    for (const line of lines) {
      const trimmed = line.trim();
      if (!funcLine && trimmed.match(/(async\s+)?func\s+\w/)) {
        funcLine = trimmed;
        inParams = trimmed.includes('(');
        parenCount = (trimmed.match(/\(/g) || []).length - (trimmed.match(/\)/g) || []).length;
        if (parenCount === 0 && trimmed.includes(')')) {
          // Single line function
          match = trimmed.match(funcRegex);
          break;
        }
      } else if (funcLine && inParams) {
        funcLine += ' ' + trimmed;
        parenCount += (trimmed.match(/\(/g) || []).length - (trimmed.match(/\)/g) || []).length;
        if (parenCount === 0 && trimmed.includes(')')) {
          // Found closing paren
          match = funcLine.match(funcRegex);
          break;
        }
      }
    }
  }
  
  if (!match) {
    return null;
  }

  const isAsync = !!match[1];
  const name = match[2];
  const paramsStr = (match[3] || '').trim();
  const returnType = (match[4] || '').trim() || 'void';

  // Parse parameters - handle union types, generics, and complex types
  const parameters: Array<{ name: string; type: string }> = [];
  if (paramsStr) {
    // Split by comma, but be careful with nested generics/unions
    const paramParts: string[] = [];
    let currentParam = '';
    let depth = 0;
    let inString = false;
    
    for (let i = 0; i < paramsStr.length; i++) {
      const char = paramsStr[i];
      if (char === '"' || char === "'") {
        inString = !inString;
        currentParam += char;
      } else if (!inString) {
        if (char === '<' || char === '[' || char === '(') {
          depth++;
          currentParam += char;
        } else if (char === '>' || char === ']' || char === ')') {
          depth--;
          currentParam += char;
        } else if (char === ',' && depth === 0) {
          paramParts.push(currentParam.trim());
          currentParam = '';
        } else {
          currentParam += char;
        }
      } else {
        currentParam += char;
      }
    }
    if (currentParam.trim()) {
      paramParts.push(currentParam.trim());
    }
    
    for (const param of paramParts) {
      const trimmed = param.trim();
      if (!trimmed) continue;
      
      // Handle parameter with type: "name:type" or "name: type"
      // Support complex types like "name: string | int" or "name: array<int>"
      const paramMatch = trimmed.match(/^(\w+)\s*:\s*(.+)$/);
      if (paramMatch) {
        parameters.push({
          name: paramMatch[1],
          type: paramMatch[2].trim(),
        });
      } else {
        // Parameter without explicit type (shouldn't happen in OmniLang, but handle gracefully)
        parameters.push({
          name: trimmed,
          type: 'unknown',
        });
      }
    }
  }

  // Extract documentation (preceding comments)
  let documentation: string | undefined;
  const lines = text.split('\n');
  const funcLineIndex = lines.findIndex(line => line.includes(`func ${name}`));
  if (funcLineIndex > 0) {
    const docLines: string[] = [];
    for (let i = funcLineIndex - 1; i >= 0; i--) {
      const line = lines[i].trim();
      if (line.startsWith('//')) {
        docLines.unshift(line.replace(/^\/\/\s*/, ''));
      } else if (line === '' || line.startsWith('/*')) {
        continue;
      } else {
        break;
      }
    }
    if (docLines.length > 0) {
      documentation = docLines.join(' ');
    }
  }

  const fullName = fullModuleName ? `${fullModuleName}.${name}` : name;

  return {
    name,
    fullName,
    module: moduleName,
    parameters,
    returnType,
    documentation,
    isAsync,
  };
}

// Parse a single std library file
function parseStdFile(filePath: string, moduleName: string, fullModuleName: string): StdFunction[] {
  const functions: StdFunction[] = [];
  
  try {
    const content = fs.readFileSync(filePath, 'utf-8');
    
    // Split by function declarations
    // Look for function declarations (may span multiple lines)
    const lines = content.split('\n');
    let currentFunc = '';
    let inFunction = false;
    let braceCount = 0;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      
      // Check if this line starts a function
      if (line.match(/(async\s+)?func\s+\w/)) {
        if (inFunction) {
          // Parse previous function
          const func = parseFunctionSignature(currentFunc, moduleName, fullModuleName);
          if (func) {
            functions.push(func);
          }
        }
        currentFunc = line;
        inFunction = true;
        braceCount = (line.match(/\{/g) || []).length - (line.match(/\}/g) || []).length;
      } else if (inFunction) {
        currentFunc += '\n' + line;
        braceCount += (line.match(/\{/g) || []).length - (line.match(/\}/g) || []).length;
        
        if (braceCount <= 0 && line.includes('}')) {
          // Function complete
          const func = parseFunctionSignature(currentFunc, moduleName, fullModuleName);
          if (func) {
            functions.push(func);
          }
          currentFunc = '';
          inFunction = false;
          braceCount = 0;
        }
      }
    }

    // Handle last function if file ends without closing brace
    if (inFunction && currentFunc) {
      const func = parseFunctionSignature(currentFunc, moduleName, fullModuleName);
      if (func) {
        functions.push(func);
      }
    }
  } catch (error) {
    // Error will be logged by caller if connection is available
    throw error;
  }

  return functions;
}

// Parse standard library directory
export async function parseStandardLibrary(
  connection: Connection | null = null,
  workspaceFolders: WorkspaceFolder[] | null = null
): Promise<StdLibrary> {
  const stdDir = findStdDirectory(connection, workspaceFolders);
  
  if (!stdDir) {
    if (connection) {
      connection.console.warn('Standard library directory not found - autocomplete for std functions will not work');
    }
    return {
      modules: new Map(),
      functions: new Map(),
    };
  }

  if (connection) {
    connection.console.log(`Parsing standard library from: ${stdDir}`);
  }

  const modules = new Map<string, StdModule>();
  const functions = new Map<string, StdFunction>();

  // Known std modules
  const moduleDirs = [
    'io', 'math', 'string', 'array', 'collections', 'file', 'os',
    'log', 'time', 'network', 'dev', 'test', 'testing', 'algorithms'
  ];

  // Parse std.omni (main module)
  const stdOmniPath = path.join(stdDir, 'std.omni');
  if (fs.existsSync(stdOmniPath)) {
    const stdFunctions = parseStdFile(stdOmniPath, 'std', 'std');
    const stdModule: StdModule = {
      name: 'std',
      functions: stdFunctions,
    };
    modules.set('std', stdModule);
    stdFunctions.forEach(func => {
      functions.set(func.fullName, func);
    });
  }

  // Parse each module directory
  for (const moduleDir of moduleDirs) {
    const modulePath = path.join(stdDir, moduleDir);
    if (!fs.existsSync(modulePath) || !fs.statSync(modulePath).isDirectory()) {
      continue;
    }

    // Find .omni files in module directory
    const files = fs.readdirSync(modulePath).filter(f => f.endsWith('.omni'));
    
    const moduleFunctions: StdFunction[] = [];
    
    for (const file of files) {
      const filePath = path.join(modulePath, file);
      const fullModuleName = `std.${moduleDir}`;
      try {
        const fileFunctions = parseStdFile(filePath, moduleDir, fullModuleName);
        moduleFunctions.push(...fileFunctions);
      } catch (error) {
        if (connection) {
          connection.console.error(`Failed to parse ${filePath}: ${error}`);
        }
        // Continue with other files
      }
    }

    const module: StdModule = {
      name: moduleDir,
      functions: moduleFunctions,
    };

    modules.set(moduleDir, module);
    modules.set(`std.${moduleDir}`, module);

    moduleFunctions.forEach(func => {
      functions.set(func.fullName, func);
      // Also index by short name for module-scoped access
      functions.set(`${moduleDir}.${func.name}`, func);
    });
  }

  if (connection) {
    connection.console.log(
      `Standard library parsed successfully: ${modules.size} modules, ${functions.size} functions`
    );
  }

  return { modules, functions };
}

