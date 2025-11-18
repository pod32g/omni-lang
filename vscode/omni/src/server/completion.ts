import {
  CompletionItem,
  CompletionItemKind,
  CompletionParams,
  TextDocument,
} from 'vscode-languageserver';
import { TextDocumentPositionParams } from 'vscode-languageserver-protocol';
import { StdLibrary, StdFunction } from './stdlib';
import { SymbolTable } from './parser';

const KEYWORDS = [
  'func', 'struct', 'enum', 'type', 'import', 'let', 'var', 'const',
  'if', 'else', 'for', 'while', 'return', 'break', 'continue',
  'match', 'switch', 'case', 'default', 'try', 'catch', 'finally',
  'throw', 'defer', 'async', 'await'
];

const TYPE_KEYWORDS = [
  'int', 'long', 'float', 'double', 'bool', 'string', 'char', 'byte', 'void'
];

const STD_MODULES = [
  'std', 'std.io', 'std.math', 'std.string', 'std.array', 'std.collections',
  'std.file', 'std.os', 'std.log', 'std.time', 'std.network', 'std.dev',
  'std.test', 'std.testing', 'std.algorithms'
];

export function handleCompletion(
  params: CompletionParams,
  document: TextDocument,
  stdLibrary: StdLibrary | null,
  symbols: SymbolTable
): CompletionItem[] {
  const items: CompletionItem[] = [];
  const position = params.position;
  const line = document.getText({
    start: { line: position.line, character: 0 },
    end: position,
  });

  // Get the word at cursor position
  const wordMatch = line.match(/(\w+)$/);
  const wordBefore = wordMatch ? wordMatch[1] : '';
  const textBefore = line.substring(0, line.length - wordBefore.length).trim();

  // Context-aware completions
  if (textBefore.endsWith('std.')) {
    // After std. - show modules
    for (const moduleName of STD_MODULES) {
      if (moduleName.startsWith('std.')) {
        const shortName = moduleName.substring(4); // Remove "std."
        const item: CompletionItem = {
          label: shortName,
          kind: CompletionItemKind.Module,
          detail: `std.${shortName}`,
          documentation: `Standard library module: ${shortName}`,
        };
        items.push(item);
      }
    }
    return items;
  }

  // After std.io. or similar - show functions in that module
  const moduleMatch = textBefore.match(/std\.(\w+)\.$/);
  if (moduleMatch) {
    const moduleName = moduleMatch[1];
    if (stdLibrary && stdLibrary.modules) {
      // Try both "io" and "std.io" keys
      let module = stdLibrary.modules.get(moduleName);
      if (!module) {
        module = stdLibrary.modules.get(`std.${moduleName}`);
      }
      
      if (module && module.functions && module.functions.length > 0) {
        for (const func of module.functions) {
          const item: CompletionItem = {
            label: func.name,
            kind: CompletionItemKind.Function,
            detail: formatFunctionSignature(func),
            documentation: func.documentation || `Function: ${func.fullName}`,
            insertText: func.name + '(',
          };
          items.push(item);
        }
      }
    }
    return items;
  }

  // After import - show modules
  if (textBefore.match(/^import\s*$/)) {
    for (const moduleName of STD_MODULES) {
      const item: CompletionItem = {
        label: moduleName,
        kind: CompletionItemKind.Module,
        documentation: `Import ${moduleName} module`,
      };
      items.push(item);
    }
    return items;
  }

  // After . - show struct/enum fields or methods
  if (textBefore.endsWith('.')) {
    const beforeDot = textBefore.slice(0, -1).trim();
    // Try to find the variable/struct type
    // This is simplified - in a full implementation, we'd do type inference
    for (const [name, struct] of symbols.structs) {
      if (beforeDot.includes(name)) {
        for (const field of struct.fields) {
          const item: CompletionItem = {
            label: field.name,
            kind: CompletionItemKind.Field,
            detail: field.type,
            documentation: `Field: ${field.name}: ${field.type}`,
          };
          items.push(item);
        }
      }
    }
    return items;
  }

  // After : - show types
  if (textBefore.endsWith(':')) {
    for (const type of TYPE_KEYWORDS) {
      items.push({
        label: type,
        kind: CompletionItemKind.TypeParameter,
      });
    }
    for (const [name, typeSymbol] of symbols.types) {
      items.push({
        label: name,
        kind: CompletionItemKind.TypeParameter,
        detail: typeSymbol.aliasedType,
      });
    }
    for (const [name] of symbols.structs) {
      items.push({
        label: name,
        kind: CompletionItemKind.Class,
      });
    }
    for (const [name] of symbols.enums) {
      items.push({
        label: name,
        kind: CompletionItemKind.Enum,
      });
    }
    return items;
  }

  // Default completions: keywords, types, std modules, local symbols
  for (const keyword of KEYWORDS) {
    if (!wordBefore || keyword.startsWith(wordBefore)) {
      items.push({
        label: keyword,
        kind: CompletionItemKind.Keyword,
      });
    }
  }

  for (const type of TYPE_KEYWORDS) {
    if (!wordBefore || type.startsWith(wordBefore)) {
      items.push({
        label: type,
        kind: CompletionItemKind.TypeParameter,
      });
    }
  }

  // Standard library modules
  for (const moduleName of STD_MODULES) {
    if (!wordBefore || moduleName.startsWith(wordBefore)) {
      items.push({
        label: moduleName,
        kind: CompletionItemKind.Module,
      });
    }
  }

  // Standard library functions (if no prefix filter)
  // Only show if stdLibrary is properly initialized and has functions
  if (stdLibrary && stdLibrary.functions && stdLibrary.functions.size > 0) {
    // Only show std functions if wordBefore is short or empty (to avoid too many completions)
    if (!wordBefore || wordBefore.length < 2) {
      for (const [fullName, func] of stdLibrary.functions) {
        if (!wordBefore || func.name.startsWith(wordBefore)) {
          items.push({
            label: func.name,
            kind: CompletionItemKind.Function,
            detail: formatFunctionSignature(func),
            documentation: func.documentation || `Function: ${func.fullName}`,
            insertText: func.name + '(',
          });
        }
      }
    }
  }

  // Local symbols
  for (const [name, func] of symbols.functions) {
    if (!wordBefore || name.startsWith(wordBefore)) {
      items.push({
        label: name,
        kind: CompletionItemKind.Function,
        detail: formatLocalFunctionSignature(func),
        insertText: name + '(',
      });
    }
  }

  for (const [name, variable] of symbols.variables) {
    if (!wordBefore || name.startsWith(wordBefore)) {
      items.push({
        label: name,
        kind: CompletionItemKind.Variable,
        detail: variable.type,
      });
    }
  }

  for (const [name] of symbols.structs) {
    if (!wordBefore || name.startsWith(wordBefore)) {
      items.push({
        label: name,
        kind: CompletionItemKind.Class,
      });
    }
  }

  for (const [name] of symbols.enums) {
    if (!wordBefore || name.startsWith(wordBefore)) {
      items.push({
        label: name,
        kind: CompletionItemKind.Enum,
      });
    }
  }

  return items;
}

function formatFunctionSignature(func: StdFunction): string {
  const params = func.parameters.map(p => `${p.name}: ${p.type}`).join(', ');
  const asyncPrefix = func.isAsync ? 'async ' : '';
  return `${asyncPrefix}func ${func.name}(${params}): ${func.returnType}`;
}

function formatLocalFunctionSignature(func: any): string {
  const params = func.parameters.map((p: any) => `${p.name}: ${p.type}`).join(', ');
  const asyncPrefix = func.isAsync ? 'async ' : '';
  return `${asyncPrefix}func ${func.name}(${params}): ${func.returnType}`;
}

