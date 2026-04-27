import {
  Definition,
  DefinitionParams,
  Location,
  TextDocument,
} from 'vscode-languageserver';
import { TextDocuments as TextDocumentsType } from 'vscode-languageserver/node';
import { SymbolTable } from './parser';

export function handleDefinition(
  params: DefinitionParams,
  document: TextDocument,
  symbols: SymbolTable,
  documents?: TextDocumentsType<TextDocument>,
  symbolTables?: Map<string, SymbolTable>
): Definition | null {
  const word = getWordAtPosition(document, params.position.line, params.position.character);
  if (!word) {
    return null;
  }

  const localDefinition = findDefinitionInSymbols(document.uri, word, symbols);
  if (localDefinition) {
    return localDefinition;
  }

  if (!documents || !symbolTables) {
    return null;
  }

  for (const candidate of documents.all()) {
    if (candidate.uri === document.uri) {
      continue;
    }

    const candidateSymbols = symbolTables.get(candidate.uri);
    if (!candidateSymbols) {
      continue;
    }

    const definition = findDefinitionInSymbols(candidate.uri, word, candidateSymbols);
    if (definition) {
      return definition;
    }
  }

  return null;
}

function findDefinitionInSymbols(
  uri: string,
  word: string,
  symbols: SymbolTable
): Location | null {
  const func = symbols.functions.get(word);
  if (func) {
    return Location.create(uri, {
      start: { line: func.line, character: func.column },
      end: { line: func.line, character: func.column + func.name.length },
    });
  }

  const variable = symbols.variables.get(word);
  if (variable) {
    return Location.create(uri, {
      start: { line: variable.line, character: variable.column },
      end: { line: variable.line, character: variable.column + variable.name.length },
    });
  }

  const struct = symbols.structs.get(word);
  if (struct) {
    return Location.create(uri, {
      start: { line: struct.line, character: struct.column },
      end: { line: struct.line, character: struct.column + struct.name.length },
    });
  }

  const enumSymbol = symbols.enums.get(word);
  if (enumSymbol) {
    return Location.create(uri, {
      start: { line: enumSymbol.line, character: enumSymbol.column },
      end: { line: enumSymbol.line, character: enumSymbol.column + enumSymbol.name.length },
    });
  }

  const typeSymbol = symbols.types.get(word);
  if (typeSymbol) {
    return Location.create(uri, {
      start: { line: typeSymbol.line, character: typeSymbol.column },
      end: { line: typeSymbol.line, character: typeSymbol.column + typeSymbol.name.length },
    });
  }

  return null;
}

function getWordAtPosition(
  document: TextDocument,
  lineNumber: number,
  character: number
): string {
  const line = document.getText({
    start: { line: lineNumber, character: 0 },
    end: { line: lineNumber, character: 1000 },
  });

  let start = character;
  let end = character;

  while (start > 0 && /\w/.test(line[start - 1])) {
    start--;
  }

  while (end < line.length && /\w/.test(line[end])) {
    end++;
  }

  return start < end ? line.substring(start, end) : '';
}
