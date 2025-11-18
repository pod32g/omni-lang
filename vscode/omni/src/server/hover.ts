import {
  Hover,
  HoverParams,
  MarkupContent,
  MarkupKind,
  TextDocument,
} from 'vscode-languageserver';
import { StdLibrary, StdFunction } from './stdlib';
import { SymbolTable } from './parser';

export function handleHover(
  params: HoverParams,
  document: TextDocument,
  stdLibrary: StdLibrary | null,
  symbols: SymbolTable
): Hover | null {
  const position = params.position;
  const line = document.getText({
    start: { line: position.line, character: 0 },
    end: { line: position.line, character: 1000 },
  });

  // Get word at cursor position
  let word = '';
  let start = position.character;
  let end = position.character;
  
  // Find start of word
  while (start > 0 && /[\w]/.test(line[start - 1])) {
    start--;
  }
  
  // Find end of word
  while (end < line.length && /[\w]/.test(line[end])) {
    end++;
  }
  
  if (start < end) {
    word = line.substring(start, end);
  }
  
  if (!word) {
    return null;
  }
  const lineText = document.getText({
    start: { line: position.line, character: 0 },
    end: { line: position.line, character: 1000 },
  });

  // Check if it's a standard library function
  if (stdLibrary) {
    // Check for std.module.function pattern
    const stdFuncMatch = lineText.match(new RegExp(`std\\.(\\w+)\\.${word}`));
    if (stdFuncMatch) {
      const moduleName = stdFuncMatch[1];
      const fullName = `std.${moduleName}.${word}`;
      const func = stdLibrary.functions.get(fullName);
      if (func) {
        return createFunctionHover(func);
      }
    }

    // Check for direct function name (might be from std)
    const func = stdLibrary.functions.get(word) || stdLibrary.functions.get(`std.${word}`);
    if (func) {
      return createFunctionHover(func);
    }
  }

  // Check local functions
  const localFunc = symbols.functions.get(word);
  if (localFunc) {
    const params = localFunc.parameters.map(p => `${p.name}: ${p.type}`).join(', ');
    const asyncPrefix = localFunc.isAsync ? 'async ' : '';
    const signature = `${asyncPrefix}func ${word}(${params}): ${localFunc.returnType}`;
    
    const content: MarkupContent = {
      kind: MarkupKind.Markdown,
      value: `**Function**\n\n\`\`\`omni\n${signature}\n\`\`\``,
    };
    return { contents: content };
  }

  // Check local variables
  const variable = symbols.variables.get(word);
  if (variable) {
    const content: MarkupContent = {
      kind: MarkupKind.Markdown,
      value: `**Variable**\n\nType: \`${variable.type}\``,
    };
    return { contents: content };
  }

  // Check structs
  const struct = symbols.structs.get(word);
  if (struct) {
    const fields = struct.fields.map(f => `  ${f.name}: ${f.type}`).join('\n');
    const content: MarkupContent = {
      kind: MarkupKind.Markdown,
      value: `**Struct**\n\n\`\`\`omni\nstruct ${word} {\n${fields}\n}\n\`\`\``,
    };
    return { contents: content };
  }

  // Check enums
  const enumSymbol = symbols.enums.get(word);
  if (enumSymbol) {
    const variants = enumSymbol.variants.join(', ');
    const content: MarkupContent = {
      kind: MarkupKind.Markdown,
      value: `**Enum**\n\n\`\`\`omni\nenum ${word} {\n  ${variants}\n}\n\`\`\``,
    };
    return { contents: content };
  }

  // Check types
  const typeSymbol = symbols.types.get(word);
  if (typeSymbol) {
    const content: MarkupContent = {
      kind: MarkupKind.Markdown,
      value: `**Type Alias**\n\n\`\`\`omni\ntype ${word} = ${typeSymbol.aliasedType}\n\`\`\``,
    };
    return { contents: content };
  }

  return null;
}

function createFunctionHover(func: StdFunction): Hover {
  const params = func.parameters.map(p => `${p.name}: ${p.type}`).join(', ');
  const asyncPrefix = func.isAsync ? 'async ' : '';
  const signature = `${asyncPrefix}func ${func.name}(${params}): ${func.returnType}`;
  
  let value = `**Function**\n\n\`\`\`omni\n${signature}\n\`\`\``;
  if (func.documentation) {
    value += `\n\n${func.documentation}`;
  }
  value += `\n\nModule: \`${func.fullName}\``;

  const content: MarkupContent = {
    kind: MarkupKind.Markdown,
    value,
  };
  
  return { contents: content };
}

