import {
  ReferenceParams,
  Location,
  TextDocument,
} from 'vscode-languageserver';
import { TextDocuments as TextDocumentsType } from 'vscode-languageserver/node';
import { SymbolTable } from './parser';

export function handleReferences(
  params: ReferenceParams,
  document: TextDocument,
  symbols: SymbolTable,
  documents: TextDocumentsType<TextDocument>,
  symbolTables: Map<string, SymbolTable>
): Location[] {
  const word = getWordAtPosition(document, params.position.line, params.position.character);
  if (!word) {
    return [];
  }

  const references: Location[] = [];

  for (const candidate of documents.all()) {
    const candidateSymbols =
      symbolTables.get(candidate.uri) ||
      (candidate.uri === document.uri ? symbols : new SymbolTable());
    collectDocumentReferences(
      candidate,
      candidateSymbols,
      word,
      params.context.includeDeclaration,
      references
    );
  }

  return references;
}

function collectDocumentReferences(
  document: TextDocument,
  symbols: SymbolTable,
  word: string,
  includeDeclaration: boolean,
  references: Location[]
): void {
  const text = document.getText();
  const lines = text.split('\n');

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const regex = new RegExp(`\\b${escapeRegExp(word)}\\b`, 'g');
    let match: RegExpExecArray | null;

    while ((match = regex.exec(line)) !== null) {
      if (includeDeclaration || !isDefinition(symbols, word, i, match.index)) {
        references.push(
          Location.create(
            document.uri,
            {
              start: { line: i, character: match.index },
              end: { line: i, character: match.index + word.length },
            }
          )
        );
      }
    }
  }
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

function isDefinition(
  symbols: SymbolTable,
  word: string,
  line: number,
  column: number
): boolean {
  const func = symbols.functions.get(word);
  const variable = symbols.variables.get(word);
  const struct = symbols.structs.get(word);
  const enumSymbol = symbols.enums.get(word);
  const typeSymbol = symbols.types.get(word);

  return Boolean(
    (func && line === func.line && column === func.column) ||
      (variable && line === variable.line && column === variable.column) ||
      (struct && line === struct.line && column === struct.column) ||
      (enumSymbol && line === enumSymbol.line && column === enumSymbol.column) ||
      (typeSymbol && line === typeSymbol.line && column === typeSymbol.column)
  );
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
