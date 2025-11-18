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
  documents: TextDocumentsType<TextDocument>
): Location[] {
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
    return [];
  }
  const references: Location[] = [];

  // Search in current document
  const text = document.getText();
  const lines = text.split('\n');
  
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const regex = new RegExp(`\\b${word}\\b`, 'g');
    let match;
    
    while ((match = regex.exec(line)) !== null) {
      // Skip if it's the definition itself
      const func = symbols.functions.get(word);
      const variable = symbols.variables.get(word);
      const struct = symbols.structs.get(word);
      const enumSymbol = symbols.enums.get(word);
      const typeSymbol = symbols.types.get(word);
      
      const isDefinition = 
        (func && i === func.line && match.index === func.column) ||
        (variable && i === variable.line && match.index === variable.column) ||
        (struct && i === struct.line && match.index === struct.column) ||
        (enumSymbol && i === enumSymbol.line && match.index === enumSymbol.column) ||
        (typeSymbol && i === typeSymbol.line && match.index === typeSymbol.column);
      
      if (!isDefinition || params.context.includeDeclaration) {
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

  // TODO: Search in other documents in workspace
  // For now, only search in current document

  return references;
}

