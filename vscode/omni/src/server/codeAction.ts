import {
  CodeActionParams,
  CodeAction,
  CodeActionKind,
  TextDocument,
  TextEdit,
} from 'vscode-languageserver';

export function handleCodeAction(
  params: CodeActionParams,
  document: TextDocument
): CodeAction[] {
  const actions: CodeAction[] = [];
  const text = document.getText();
  const lines = text.split('\n');
  const importLines = collectImportLines(lines);

  if (usesStdNamespace(text) && !hasStdImport(importLines)) {
    actions.push({
      title: 'Add import std',
      kind: CodeActionKind.QuickFix,
      diagnostics: params.context.diagnostics,
      edit: {
        changes: {
          [document.uri]: [TextEdit.insert(importInsertionPosition(lines), 'import std\n')],
        },
      },
    });
  }

  const duplicateImportEdits = duplicateImportRemovalEdits(importLines);
  if (duplicateImportEdits.length > 0) {
    actions.push({
      title: 'Remove duplicate imports',
      kind: CodeActionKind.QuickFix,
      edit: {
        changes: {
          [document.uri]: duplicateImportEdits,
        },
      },
    });
  }

  const organizeEdit = organizeImportsEdit(importLines, lines);
  if (organizeEdit) {
    actions.push({
      title: 'Organize imports',
      kind: CodeActionKind.SourceOrganizeImports,
      edit: {
        changes: {
          [document.uri]: [organizeEdit],
        },
      },
    });
  }

  return actions;
}

interface ImportLine {
  line: number;
  text: string;
  normalized: string;
}

function collectImportLines(lines: string[]): ImportLine[] {
  const imports: ImportLine[] = [];

  for (let lineNumber = 0; lineNumber < lines.length; lineNumber++) {
    const text = lines[lineNumber];
    const trimmed = text.trim();
    if (!trimmed.startsWith('import ')) {
      continue;
    }

    imports.push({
      line: lineNumber,
      text: trimmed,
      normalized: trimmed.replace(/\s+/g, ' '),
    });
  }

  return imports;
}

function usesStdNamespace(text: string): boolean {
  return /\bstd\.[A-Za-z_]\w*/.test(stripComments(text));
}

function hasStdImport(imports: ImportLine[]): boolean {
  return imports.some(item => item.normalized === 'import std');
}

function importInsertionPosition(lines: string[]) {
  let line = 0;

  while (line < lines.length) {
    const trimmed = lines[line].trim();
    if (trimmed === '' || trimmed.startsWith('//')) {
      line++;
      continue;
    }
    break;
  }

  return { line, character: 0 };
}

function duplicateImportRemovalEdits(imports: ImportLine[]): TextEdit[] {
  const seen = new Set<string>();
  const edits: TextEdit[] = [];

  for (const item of imports) {
    if (!seen.has(item.normalized)) {
      seen.add(item.normalized);
      continue;
    }

    edits.push(
      TextEdit.del({
        start: { line: item.line, character: 0 },
        end: { line: item.line + 1, character: 0 },
      })
    );
  }

  return edits;
}

function organizeImportsEdit(imports: ImportLine[], lines: string[]): TextEdit | null {
  if (imports.length < 2) {
    return null;
  }

  const firstLine = imports[0].line;
  const lastLine = imports[imports.length - 1].line;

  for (let line = firstLine; line <= lastLine; line++) {
    const trimmed = lines[line].trim();
    if (trimmed !== '' && !trimmed.startsWith('import ')) {
      return null;
    }
  }

  const organized = Array.from(new Set(imports.map(item => item.normalized))).sort();
  const current = imports.map(item => item.normalized);
  if (
    organized.length === current.length &&
    organized.every((value, index) => value === current[index])
  ) {
    return null;
  }

  return TextEdit.replace(
    {
      start: { line: firstLine, character: 0 },
      end: { line: lastLine + 1, character: 0 },
    },
    organized.join('\n') + '\n'
  );
}

function stripComments(text: string): string {
  return text
    .split('\n')
    .map(line => {
      const commentStart = line.indexOf('//');
      return commentStart >= 0 ? line.slice(0, commentStart) : line;
    })
    .join('\n');
}
