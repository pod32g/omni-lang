import {
  Definition,
  DefinitionParams,
  Location,
  TextDocument,
} from 'vscode-languageserver';
import { SymbolTable } from './parser';

export function handleDefinition(
  params: DefinitionParams,
  document: TextDocument,
  symbols: SymbolTable
): Definition | null {
  const position = params.position;
  
  // Get word at position manually
  const line = document.getText({
    start: { line: position.line, character: 0 },
    end: { line: position.line, character: 1000 },
  });
  
  // Find word at cursor position
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

  // Check functions
  const func = symbols.functions.get(word);
  if (func) {
    return Location.create(
      document.uri,
      {
        start: { line: func.line, character: func.column },
        end: { line: func.line, character: func.column + func.name.length },
      }
    );
  }

  // Check variables
  const variable = symbols.variables.get(word);
  if (variable) {
    return Location.create(
      document.uri,
      {
        start: { line: variable.line, character: variable.column },
        end: { line: variable.line, character: variable.column + variable.name.length },
      }
    );
  }

  // Check structs
  const struct = symbols.structs.get(word);
  if (struct) {
    return Location.create(
      document.uri,
      {
        start: { line: struct.line, character: struct.column },
        end: { line: struct.line, character: struct.column + struct.name.length },
      }
    );
  }

  // Check enums
  const enumSymbol = symbols.enums.get(word);
  if (enumSymbol) {
    return Location.create(
      document.uri,
      {
        start: { line: enumSymbol.line, character: enumSymbol.column },
        end: { line: enumSymbol.line, character: enumSymbol.column + enumSymbol.name.length },
      }
    );
  }

  // Check types
  const typeSymbol = symbols.types.get(word);
  if (typeSymbol) {
    return Location.create(
      document.uri,
      {
        start: { line: typeSymbol.line, character: typeSymbol.column },
        end: { line: typeSymbol.line, character: typeSymbol.column + typeSymbol.name.length },
      }
    );
  }

  return null;
}

