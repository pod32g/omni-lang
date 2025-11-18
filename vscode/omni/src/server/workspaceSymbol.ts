import {
  WorkspaceSymbolParams,
  SymbolInformation,
  SymbolKind,
  TextDocuments,
} from 'vscode-languageserver';
import { TextDocuments as TextDocumentsType } from 'vscode-languageserver/node';
import { TextDocument } from 'vscode-languageserver-textdocument';
import { SymbolTable } from './parser';

export function handleWorkspaceSymbol(
  params: WorkspaceSymbolParams,
  symbolTables: Map<string, SymbolTable>,
  documents: TextDocumentsType<TextDocument>
): SymbolInformation[] {
  const query = params.query.toLowerCase();
  const results: SymbolInformation[] = [];

  for (const [uri, symbols] of symbolTables) {
    const document = documents.get(uri);
    if (!document) continue;

    // Search functions
    for (const [name, func] of symbols.functions) {
      if (name.toLowerCase().includes(query)) {
        results.push({
          name,
          kind: SymbolKind.Function,
          location: {
            uri,
            range: {
              start: { line: func.line, character: func.column },
              end: { line: func.line, character: func.column + name.length },
            },
          },
        });
      }
    }

    // Search structs
    for (const [name, struct] of symbols.structs) {
      if (name.toLowerCase().includes(query)) {
        results.push({
          name,
          kind: SymbolKind.Struct,
          location: {
            uri,
            range: {
              start: { line: struct.line, character: struct.column },
              end: { line: struct.line, character: struct.column + name.length },
            },
          },
        });
      }
    }

    // Search enums
    for (const [name, enumSymbol] of symbols.enums) {
      if (name.toLowerCase().includes(query)) {
        results.push({
          name,
          kind: SymbolKind.Enum,
          location: {
            uri,
            range: {
              start: { line: enumSymbol.line, character: enumSymbol.column },
              end: { line: enumSymbol.line, character: enumSymbol.column + name.length },
            },
          },
        });
      }
    }
  }

  return results;
}

