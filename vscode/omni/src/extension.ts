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
  'std.network'
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
        ['-emit', 'mir', document.isUntitled ? '-' : filePath],
        {
          cwd: vscode.workspace.rootPath ?? undefined,
          timeout: 15000
        },
        (error, _stdout, stderr) => {
          if (error && (stderr.length === 0 || !/error/i.test(stderr))) {
            reject(error);
            return;
          }
          applyDiagnostics(document, stderr);
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

function applyDiagnostics(document: vscode.TextDocument, stderr: string) {
  const diagnostics: vscode.Diagnostic[] = [];
  const lines = stderr.split(/\r?\n/);
  const regex = /(.+):(\d+):(\d+):\s*(error|warning):\s*(.+)/;

  for (const line of lines) {
    const match = regex.exec(line.trim());
    if (!match) {
      continue;
    }
    const [, file, lineStr, colStr, severityStr, message] = match;

    if (!document.fileName.endsWith(file.trim()) && file.trim() !== document.fileName) {
      continue;
    }

    const lineNum = Math.max(0, parseInt(lineStr, 10) - 1);
    const colNum = Math.max(0, parseInt(colStr, 10) - 1);
    const range = new vscode.Range(lineNum, colNum, lineNum, colNum + 1);
    const severity =
      severityStr === 'warning' ? vscode.DiagnosticSeverity.Warning : vscode.DiagnosticSeverity.Error;

    diagnostics.push(
      new vscode.Diagnostic(range, message.trim(), severity)
    );
  }

  diagnosticCollection.set(document.uri, diagnostics);
}

async function writeTempDocument(document: vscode.TextDocument): Promise<string> {
  const tempFile = join(tmpdir(), `omni-${Date.now()}-${Math.random().toString(16).slice(2)}.omni`);
  await fs.writeFile(tempFile, document.getText(), 'utf8');
  return tempFile;
}

