import {
  SignatureHelp,
  SignatureHelpParams,
  SignatureInformation,
  ParameterInformation,
  TextDocument,
} from 'vscode-languageserver';
import { StdLibrary, StdFunction } from './stdlib';
import { SymbolTable } from './parser';

export function handleSignatureHelp(
  params: SignatureHelpParams,
  document: TextDocument,
  stdLibrary: StdLibrary | null,
  symbols: SymbolTable
): SignatureHelp | null {
  const position = params.position;
  const line = document.getText({
    start: { line: position.line, character: 0 },
    end: position,
  });

  // Find function call - look backwards for function name
  const funcMatch = line.match(/(\w+)\s*\(/);
  if (!funcMatch) {
    return null;
  }

  const funcName = funcMatch[1];
  
  // Count commas to determine active parameter
  const textAfterOpenParen = line.substring(line.lastIndexOf('(') + 1);
  const activeParameter = (textAfterOpenParen.match(/,/g) || []).length;

  // Check standard library
  if (stdLibrary) {
    // Try to find function in std library
    // Check for std.module.function pattern
    const stdFuncMatch = line.match(/std\.(\w+)\.(\w+)\s*\(/);
    if (stdFuncMatch) {
      const moduleName = stdFuncMatch[1];
      const actualFuncName = stdFuncMatch[2];
      const fullName = `std.${moduleName}.${actualFuncName}`;
      const func = stdLibrary.functions.get(fullName);
      if (func) {
        return createSignatureHelp(func, activeParameter);
      }
    }

    // Check for direct function name
    const func = stdLibrary.functions.get(funcName) || 
                 stdLibrary.functions.get(`std.${funcName}`);
    if (func) {
      return createSignatureHelp(func, activeParameter);
    }
  }

  // Check local functions
  const localFunc = symbols.functions.get(funcName);
  if (localFunc) {
    const params = localFunc.parameters.map((p, i) => {
      return ParameterInformation.create(`${p.name}: ${p.type}`);
    });

    const signature: SignatureInformation = {
      label: `${funcName}(${localFunc.parameters.map(p => `${p.name}: ${p.type}`).join(', ')})`,
      documentation: `Returns: ${localFunc.returnType}`,
      parameters: params,
    };

    return {
      signatures: [signature],
      activeSignature: 0,
      activeParameter: Math.min(activeParameter, params.length - 1),
    };
  }

  return null;
}

function createSignatureHelp(func: StdFunction, activeParameter: number): SignatureHelp {
  const params = func.parameters.map((p, i) => {
    return ParameterInformation.create(`${p.name}: ${p.type}`);
  });

  const signature: SignatureInformation = {
    label: `${func.name}(${func.parameters.map(p => `${p.name}: ${p.type}`).join(', ')})`,
    documentation: func.documentation || `Returns: ${func.returnType}`,
    parameters: params,
  };

  return {
    signatures: [signature],
    activeSignature: 0,
    activeParameter: Math.min(activeParameter, params.length - 1),
  };
}

