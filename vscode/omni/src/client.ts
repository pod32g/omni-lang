import * as vscode from 'vscode';
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from 'vscode-languageclient/node';
import * as path from 'path';

let client: LanguageClient | null = null;

export function activateLanguageClient(context: vscode.ExtensionContext): LanguageClient {
  // The server is implemented in Node
  const serverModule = context.asAbsolutePath(path.join('dist', 'server', 'index.js'));

  // The debug options for the server
  const debugOptions = { execArgv: ['--nolazy', '--inspect=6009'] };

  // If the extension is launched in debug mode then the debug server options are used
  // Otherwise the run options are used
  const serverOptions: ServerOptions = {
    run: { module: serverModule, transport: TransportKind.stdio },
    debug: {
      module: serverModule,
      transport: TransportKind.stdio,
      options: debugOptions,
    },
  };

  // Options to control the language client
  const clientOptions: LanguageClientOptions = {
    // Register the server for OmniLang documents
    documentSelector: [{ scheme: 'file', language: 'omni' }],
    synchronize: {
      // Notify the server about file changes to .omni files contained in the workspace
      fileEvents: vscode.workspace.createFileSystemWatcher('**/*.omni'),
    },
    // Pass workspace folder information
    workspaceFolder: vscode.workspace.workspaceFolders?.[0],
    // Pass initialization options
    initializationOptions: {
      workspaceRoot: vscode.workspace.workspaceFolders?.[0]?.uri.fsPath,
    },
  };

  // Create the language client and start the client
  client = new LanguageClient(
    'omniLang',
    'OmniLang Language Server',
    serverOptions,
    clientOptions
  );

  // Start the client. This will also launch the server
  client.start();

  return client;
}

export function deactivateLanguageClient(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}

