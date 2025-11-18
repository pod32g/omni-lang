import * as fs from 'fs';
import * as path from 'path';

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
function findStdDirectory(): string | null {
  // Try to find std directory relative to extension
  // When compiled, __dirname will be in dist/server/
  // We need to go up to find the omni/std directory
  const possiblePaths = [
    path.join(__dirname, '../../../../omni/std'),
    path.join(__dirname, '../../../omni/std'),
    path.join(__dirname, '../../../../../omni/std'),
    path.join(process.cwd(), 'omni/std'),
    path.join(process.cwd(), '../omni/std'),
    // Try relative to workspace if available
    ...(process.env.OMNI_STD_PATH ? [process.env.OMNI_STD_PATH] : []),
  ];

  for (const stdPath of possiblePaths) {
    try {
      if (fs.existsSync(stdPath) && fs.statSync(stdPath).isDirectory()) {
        return stdPath;
      }
    } catch (e) {
      // Ignore errors
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
  // Match: async? func functionName(params):returnType
  // or: func functionName(params)
  const funcRegex = /(async\s+)?func\s+(\w+)\s*\(([^)]*)\)\s*:?\s*([^{]*?)\s*\{/;
  const match = text.match(funcRegex);
  
  if (!match) {
    return null;
  }

  const isAsync = !!match[1];
  const name = match[2];
  const paramsStr = match[3].trim();
  const returnType = match[4].trim() || 'void';

  // Parse parameters
  const parameters: Array<{ name: string; type: string }> = [];
  if (paramsStr) {
    const paramParts = paramsStr.split(',').map(p => p.trim());
    for (const param of paramParts) {
      if (!param) continue;
      // Handle parameter with type: "name:type" or "name: type"
      const paramMatch = param.match(/^(\w+)\s*:\s*(.+)$/);
      if (paramMatch) {
        parameters.push({
          name: paramMatch[1],
          type: paramMatch[2].trim(),
        });
      } else {
        // Parameter without explicit type
        parameters.push({
          name: param,
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
    console.error(`Failed to parse ${filePath}: ${error}`);
  }

  return functions;
}

// Parse standard library directory
export async function parseStandardLibrary(): Promise<StdLibrary> {
  const stdDir = findStdDirectory();
  
  if (!stdDir) {
    console.warn('Standard library directory not found');
    return {
      modules: new Map(),
      functions: new Map(),
    };
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
      const fileFunctions = parseStdFile(filePath, moduleDir, fullModuleName);
      moduleFunctions.push(...fileFunctions);
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

  return { modules, functions };
}

