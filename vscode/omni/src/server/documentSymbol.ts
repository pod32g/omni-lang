import {
  DocumentSymbolParams,
  DocumentSymbol,
  SymbolKind,
  TextDocument,
} from 'vscode-languageserver';
import { SymbolTable } from './parser';

export function handleDocumentSymbol(
  params: DocumentSymbolParams,
  document: TextDocument,
  symbols: SymbolTable
): DocumentSymbol[] {
  const result: DocumentSymbol[] = [];

  // Add functions
  for (const [name, func] of symbols.functions) {
    result.push({
      name,
      kind: SymbolKind.Function,
      range: {
        start: { line: func.line, character: func.column },
        end: { line: func.line, character: func.column + name.length },
      },
      selectionRange: {
        start: { line: func.line, character: func.column },
        end: { line: func.line, character: func.column + name.length },
      },
      detail: formatFunctionDetail(func),
    });
  }

  // Add structs
  for (const [name, struct] of symbols.structs) {
    const children: DocumentSymbol[] = struct.fields.map(field => ({
      name: field.name,
      kind: SymbolKind.Field,
      range: {
        start: { line: struct.line, character: 0 },
        end: { line: struct.line, character: 0 },
      },
      selectionRange: {
        start: { line: struct.line, character: 0 },
        end: { line: struct.line, character: 0 },
      },
      detail: field.type,
    }));

    result.push({
      name,
      kind: SymbolKind.Struct,
      range: {
        start: { line: struct.line, character: struct.column },
        end: { line: struct.line, character: struct.column + name.length },
      },
      selectionRange: {
        start: { line: struct.line, character: struct.column },
        end: { line: struct.line, character: struct.column + name.length },
      },
      children: children.length > 0 ? children : undefined,
    });
  }

  // Add enums
  for (const [name, enumSymbol] of symbols.enums) {
    const children: DocumentSymbol[] = enumSymbol.variants.map(variant => ({
      name: variant,
      kind: SymbolKind.EnumMember,
      range: {
        start: { line: enumSymbol.line, character: 0 },
        end: { line: enumSymbol.line, character: 0 },
      },
      selectionRange: {
        start: { line: enumSymbol.line, character: 0 },
        end: { line: enumSymbol.line, character: 0 },
      },
    }));

    result.push({
      name,
      kind: SymbolKind.Enum,
      range: {
        start: { line: enumSymbol.line, character: enumSymbol.column },
        end: { line: enumSymbol.line, character: enumSymbol.column + name.length },
      },
      selectionRange: {
        start: { line: enumSymbol.line, character: enumSymbol.column },
        end: { line: enumSymbol.line, character: enumSymbol.column + name.length },
      },
      children: children.length > 0 ? children : undefined,
    });
  }

  // Add variables (top-level only)
  for (const [name, variable] of symbols.variables) {
    result.push({
      name,
      kind: SymbolKind.Variable,
      range: {
        start: { line: variable.line, character: variable.column },
        end: { line: variable.line, character: variable.column + name.length },
      },
      selectionRange: {
        start: { line: variable.line, character: variable.column },
        end: { line: variable.line, character: variable.column + name.length },
      },
      detail: variable.type,
    });
  }

  // Add types
  for (const [name, typeSymbol] of symbols.types) {
    result.push({
      name,
      kind: SymbolKind.TypeParameter,
      range: {
        start: { line: typeSymbol.line, character: typeSymbol.column },
        end: { line: typeSymbol.line, character: typeSymbol.column + name.length },
      },
      selectionRange: {
        start: { line: typeSymbol.line, character: typeSymbol.column },
        end: { line: typeSymbol.line, character: typeSymbol.column + name.length },
      },
      detail: typeSymbol.aliasedType,
    });
  }

  return result;
}

function formatFunctionDetail(func: any): string {
  const params = func.parameters.map((p: any) => `${p.name}: ${p.type}`).join(', ');
  return `(${params}): ${func.returnType}`;
}

