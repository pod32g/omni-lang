import * as vscode from 'vscode';
import { activateLanguageClient, deactivateLanguageClient } from './client';

let client: any = null;

export function activate(context: vscode.ExtensionContext) {
  // Activate the language client
  client = activateLanguageClient(context);
  
  // Register compile command (keep for backward compatibility)
  context.subscriptions.push(
    vscode.commands.registerCommand('omni-lang.compileCurrentFile', async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor || editor.document.languageId !== 'omni') {
        vscode.window.showInformationMessage('No Omni file is currently active.');
        return;
      }
      vscode.window.showInformationMessage('OmniLang compilation finished (see PROBLEMS for details).');
    })
  );
}

export function deactivate(): Thenable<void> | undefined {
  return deactivateLanguageClient();
}
