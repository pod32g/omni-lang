import * as vscode from 'vscode';
import { execFile } from 'node:child_process';
import { promises as fs } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

const KEYWORDS = [
  'func',
  'struct',
  'enum',
  'type',
  'import',
  'let',
  'var',
  'const',
  'if',
  'else',
  'for',
  'while',
  'return',
  'break',
  'continue',
  'match',
  'switch',
  'case',
  'default',
  'try',
  'catch',
  'finally',
  'throw',
  'defer',
  'async',
  'await'
];

const STD_NAMESPACES = [
  'std',
  'std.io',
  'std.math',
  'std.string',
  'std.file',
  'std.array',
  'std.collections',
  'std.log',
  'std.os',
  'std.time',
  'std.network',
  'std.testing',
  'std.dev'
];

const TYPE_KEYWORDS = [
  'int',
  'long',
  'float',
  'double',
  'bool',
  'string',
  'char',
  'byte',
  'void'
];

let diagnosticCollection: vscode.DiagnosticCollection;

export function activate(context: vscode.ExtensionContext) {
  diagnosticCollection = vscode.languages.createDiagnosticCollection('omni');
  context.subscriptions.push(diagnosticCollection);

  context.subscriptions.push(
    vscode.languages.registerCompletionItemProvider(
      { language: 'omni', scheme: 'file' },
      {
        provideCompletionItems(document, position) {
          const items: vscode.CompletionItem[] = [];

          for (const keyword of KEYWORDS) {
            const item = new vscode.CompletionItem(keyword, vscode.CompletionItemKind.Keyword);
            items.push(item);
          }

          for (const type of TYPE_KEYWORDS) {
            const item = new vscode.CompletionItem(type, vscode.CompletionItemKind.TypeParameter);
            items.push(item);
          }

          for (const ns of STD_NAMESPACES) {
            const item = new vscode.CompletionItem(ns, vscode.CompletionItemKind.Module);
            items.push(item);
          }

          const customTypeMatch = document.getText().match(/\b(struct|enum)\s+([A-Z][A-Za-z0-9_]*)/g);
          if (customTypeMatch) {
            const seen = new Set<string>();
            for (const match of customTypeMatch) {
              const parts = match.split(/\s+/);
              const name = parts[1];
              if (name && !seen.has(name)) {
                seen.add(name);
                const item = new vscode.CompletionItem(name, vscode.CompletionItemKind.Class);
                items.push(item);
              }
            }
          }

          return items;
        }
      },
      '.',
      ':'
    )
  );

  context.subscriptions.push(
    vscode.languages.registerHoverProvider('omni', {
      provideHover(document, position) {
        const range = document.getWordRangeAtPosition(position);
        if (!range) {
          return;
        }
        const word = document.getText(range);

        if (KEYWORDS.includes(word)) {
          return new vscode.Hover(`**Keyword** \`${word}\``);
        }

        if (TYPE_KEYWORDS.includes(word)) {
          return new vscode.Hover(`**Primitive type** \`${word}\``);
        }

        if (STD_NAMESPACES.includes(word)) {
          return new vscode.Hover(`**Standard module** \`${word}\``);
        }

        return;
      }
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('omni-lang.compileCurrentFile', async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor || editor.document.languageId !== 'omni') {
        vscode.window.showInformationMessage('No Omni file is currently active.');
        return;
      }
      await runDiagnostics(editor.document, true);
      vscode.window.showInformationMessage('OmniLang compilation finished (see PROBLEMS for details).');
    })
  );

  context.subscriptions.push(
    vscode.workspace.onDidOpenTextDocument((doc) => {
      if (doc.languageId === 'omni') {
        runDiagnostics(doc);
      }
    })
  );

  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((doc) => {
      if (doc.languageId === 'omni') {
        runDiagnostics(doc);
      }
    })
  );

  context.subscriptions.push(
    vscode.workspace.onDidChangeTextDocument((event) => {
      if (event.document.languageId === 'omni') {
        runDiagnostics(event.document);
      }
    })
  );

  vscode.workspace.textDocuments
    .filter((doc) => doc.languageId === 'omni')
    .forEach((doc) => runDiagnostics(doc));
}

export function deactivate() {
  diagnosticCollection.dispose();
}

async function runDiagnostics(document: vscode.TextDocument, force = false) {
  const config = vscode.workspace.getConfiguration('omniLang');
  const omnicPath = config.get<string>('omnicPath', 'omnic');

  let tempFilePath: string | undefined;
  let filePath = document.fileName;

  try {
    if (document.isDirty || document.isUntitled) {
      tempFilePath = await writeTempDocument(document);
      filePath = tempFilePath;
    }

    await new Promise<void>((resolve, reject) => {
      execFile(
        omnicPath,
        ['--diagnostics-json', '-emit', 'mir', document.isUntitled ? '-' : filePath],
        {
          cwd: vscode.workspace.rootPath ?? undefined,
          timeout: 15000
        },
        (error, stdout, stderr) => {
          if (error && !isDiagnosticError(stdout, stderr)) {
            reject(error);
            return;
          }
          if (!applyJsonDiagnostics(document, stdout)) {
            if (!applyTextDiagnostics(document, stderr)) {
              diagnosticCollection.set(document.uri, []);
            }
          }
          resolve();
        }
      );
    });
  } catch (err) {
    if (force) {
      const message = err instanceof Error ? err.message : String(err);
      vscode.window.showErrorMessage(`OmniLang compilation failed: ${message}`);
    }
    diagnosticCollection.set(document.uri, []);
  } finally {
    if (tempFilePath) {
      void fs.unlink(tempFilePath);
    }
  }
}

function normalizePath(path: string): string {
  return vscode.Uri.file(path).fsPath;
}

function appliesToDocument(document: vscode.TextDocument, file: string): boolean {
  const documentPath = document.uri.fsPath;
  const normalizedFile = normalizePath(file.trim());
  return documentPath === normalizedFile || documentPath.endsWith(normalizedFile);
}

function applyJsonDiagnostics(document: vscode.TextDocument, stdout: string): boolean {
  if (!stdout.trim()) {
    return false;
  }

  const jsonLine = stdout
    .split(/\r?\n/)
    .map((line) => line.trim())
    .find((line) => line.startsWith('{') && line.endsWith('}'));

  if (!jsonLine) {
    return false;
  }

  try {
    const payload = JSON.parse(jsonLine) as {
      status?: string;
      diagnostics?: Array<{
        file: string;
        message: string;
        hint?: string;
        severity?: string;
        span?: {
          start_line: number;
          start_column: number;
          end_line: number;
          end_column: number;
        };
      }>;
    };

    if (payload.status !== 'error' || !Array.isArray(payload.diagnostics)) {
      return false;
    }

    const diagnostics: vscode.Diagnostic[] = [];

    for (const diag of payload.diagnostics) {
      if (!diag.file || !diag.span || !appliesToDocument(document, diag.file)) {
        continue;
      }

      const startLine = Math.max(0, (diag.span.start_line ?? 1) - 1);
      const startCol = Math.max(0, (diag.span.start_column ?? 1) - 1);
      const endLine = Math.max(startLine, (diag.span.end_line ?? diag.span.start_line ?? 1) - 1);
      const endCol = Math.max(
        startCol + 1,
        (diag.span.end_column ?? diag.span.start_column ?? 1) - 1
      );

      const range = new vscode.Range(startLine, startCol, endLine, endCol);
      const severityValue = (diag.severity ?? 'error').toLowerCase();
      const severity =
        severityValue === 'warning'
          ? vscode.DiagnosticSeverity.Warning
          : severityValue === 'info'
          ? vscode.DiagnosticSeverity.Information
          : vscode.DiagnosticSeverity.Error;

      const message = diag.hint ? `${diag.message}\nHint: ${diag.hint}` : diag.message;
      const diagnostic = new vscode.Diagnostic(range, message, severity);
      diagnostic.source = 'omnic';
      diagnostics.push(diagnostic);
    }

    diagnosticCollection.set(document.uri, diagnostics);
    return diagnostics.length > 0;
  } catch {
    return false;
  }
}

function applyTextDiagnostics(document: vscode.TextDocument, stderr: string): boolean {
  const diagnostics: vscode.Diagnostic[] = [];
  const lines = stderr.split(/\r?\n/);
  const regex = /(.+):(\d+):(\d+):\s*(error|warning):\s*(.+)/;

  for (const line of lines) {
    const match = regex.exec(line.trim());
    if (!match) {
      continue;
    }
    const [, file, lineStr, colStr, severityStr, message] = match;

    if (!appliesToDocument(document, file)) {
      continue;
    }

    const lineNum = Math.max(0, parseInt(lineStr, 10) - 1);
    const colNum = Math.max(0, parseInt(colStr, 10) - 1);
    const range = new vscode.Range(lineNum, colNum, lineNum, colNum + 1);
    const severity =
      severityStr === 'warning' ? vscode.DiagnosticSeverity.Warning : vscode.DiagnosticSeverity.Error;

    const diagnostic = new vscode.Diagnostic(range, message.trim(), severity);
    diagnostic.source = 'omnic';
    diagnostics.push(diagnostic);
  }

  if (diagnostics.length > 0) {
    diagnosticCollection.set(document.uri, diagnostics);
    return true;
  }

  return false;
}

function isDiagnosticError(stdout: string, stderr: string): boolean {
  if (!stdout.trim() && !stderr.trim()) {
    return false;
  }
  if (applyJsonDiagnosticsPlaceholder(stdout)) {
    return true;
  }
  return /error/i.test(stderr);
}

function applyJsonDiagnosticsPlaceholder(stdout: string): boolean {
  if (!stdout.trim()) {
    return false;
  }
  return stdout
    .split(/\r?\n/)
    .map((line) => line.trim())
    .some((line) => {
      if (!line.startsWith('{') || !line.endsWith('}')) {
        return false;
      }
      try {
        const payload = JSON.parse(line);
        return payload && typeof payload === 'object' && payload.status === 'error';
      } catch {
        return false;
      }
    });
}

async function writeTempDocument(document: vscode.TextDocument): Promise<string> {
  const tempFile = join(tmpdir(), `omni-${Date.now()}-${Math.random().toString(16).slice(2)}.omni`);
  await fs.writeFile(tempFile, document.getText(), 'utf8');
  return tempFile;
}

