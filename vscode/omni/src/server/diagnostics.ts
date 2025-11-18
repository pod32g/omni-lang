import {
  TextDocument,
  Diagnostic,
  DiagnosticSeverity,
  Connection,
} from 'vscode-languageserver';
import { execFile } from 'child_process';
import { promisify } from 'util';
import { StdLibrary } from './stdlib';

const execFileAsync = promisify(execFile);

let diagnosticTimeout: NodeJS.Timeout | null = null;

export async function handleDiagnostics(
  document: TextDocument,
  connection: Connection,
  stdLibrary: StdLibrary | null
): Promise<void> {
  // Debounce diagnostics
  if (diagnosticTimeout) {
    clearTimeout(diagnosticTimeout);
  }

  diagnosticTimeout = setTimeout(async () => {
    await runDiagnostics(document, connection);
  }, 500); // 500ms debounce
}

async function runDiagnostics(
  document: TextDocument,
  connection: Connection
): Promise<void> {
  try {
    // Try to use omnic for diagnostics
    // This requires omnic to be in PATH or configured
    const result = await execFileAsync(
      'omnic',
      ['--diagnostics-json', '-emit', 'mir', document.uri.replace('file://', '')],
      { timeout: 10000 }
    );

    const diagnostics: Diagnostic[] = [];

    // Try to parse JSON diagnostics
    if (result.stdout) {
      const jsonLine = result.stdout
        .split(/\r?\n/)
        .map(line => line.trim())
        .find(line => line.startsWith('{') && line.endsWith('}'));

      if (jsonLine) {
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

          if (payload.status === 'error' && Array.isArray(payload.diagnostics)) {
            for (const diag of payload.diagnostics) {
              if (diag.span) {
                const startLine = Math.max(0, (diag.span.start_line ?? 1) - 1);
                const startCol = Math.max(0, (diag.span.start_column ?? 1) - 1);
                const endLine = Math.max(startLine, (diag.span.end_line ?? diag.span.start_line ?? 1) - 1);
                const endCol = Math.max(
                  startCol + 1,
                  (diag.span.end_column ?? diag.span.start_column ?? 1) - 1
                );

                const severityValue = (diag.severity ?? 'error').toLowerCase();
                const severity =
                  severityValue === 'warning'
                    ? DiagnosticSeverity.Warning
                    : severityValue === 'info'
                    ? DiagnosticSeverity.Information
                    : DiagnosticSeverity.Error;

                const message = diag.hint ? `${diag.message}\nHint: ${diag.hint}` : diag.message;

                diagnostics.push({
                  range: {
                    start: { line: startLine, character: startCol },
                    end: { line: endLine, character: endCol },
                  },
                  message,
                  severity,
                  source: 'omnic',
                });
              }
            }
          }
        } catch (parseError) {
          // If JSON parsing fails, try text-based diagnostics
          parseTextDiagnostics(result.stderr, diagnostics);
        }
      } else {
        parseTextDiagnostics(result.stderr, diagnostics);
      }
    } else {
      parseTextDiagnostics(result.stderr, diagnostics);
    }

    connection.sendDiagnostics({ uri: document.uri, diagnostics });
  } catch (error) {
    // If omnic is not available, just clear diagnostics
    connection.sendDiagnostics({ uri: document.uri, diagnostics: [] });
  }
}

function parseTextDiagnostics(stderr: string, diagnostics: Diagnostic[]): void {
  const lines = stderr.split(/\r?\n/);
  const regex = /(.+):(\d+):(\d+):\s*(error|warning):\s*(.+)/;

  for (const line of lines) {
    const match = regex.exec(line.trim());
    if (!match) continue;

    const [, , lineStr, colStr, severityStr, message] = match;
    const lineNum = Math.max(0, parseInt(lineStr, 10) - 1);
    const colNum = Math.max(0, parseInt(colStr, 10) - 1);
    const severity =
      severityStr === 'warning' ? DiagnosticSeverity.Warning : DiagnosticSeverity.Error;

    diagnostics.push({
      range: {
        start: { line: lineNum, character: colNum },
        end: { line: lineNum, character: colNum + 1 },
      },
      message: message.trim(),
      severity,
      source: 'omnic',
    });
  }
}

