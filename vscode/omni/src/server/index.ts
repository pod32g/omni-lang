import {
  createConnection,
  TextDocuments,
  TextDocumentSyncKind,
  InitializeResult,
  TextDocumentChangeEvent,
} from 'vscode-languageserver/node';
import { TextDocument } from 'vscode-languageserver-textdocument';
import { parseStandardLibrary } from './stdlib';
import { handleCompletion } from './completion';
import { handleHover } from './hover';
import { handleSignatureHelp } from './signatureHelp';
import { handleDefinition } from './definition';
import { handleReferences } from './references';
import { handleDocumentSymbol } from './documentSymbol';
import { handleWorkspaceSymbol } from './workspaceSymbol';
import { handleFormatting } from './formatting';
import { handleCodeAction } from './codeAction';
import { handleDiagnostics } from './diagnostics';
import { parseDocument, SymbolTable } from './parser';

// Create a connection for the server using stdio
const connection = createConnection();

// Create a simple text document manager
const documents: TextDocuments<TextDocument> = new TextDocuments(TextDocument);

// Standard library cache
let stdLibrary: any = null;

// Document symbol tables
const symbolTables = new Map<string, SymbolTable>();

// Initialize standard library on server startup
let stdLibraryInitialized = false;

async function initializeStdLibrary(workspaceFolders: any[] | null = null) {
  if (stdLibraryInitialized) {
    return stdLibrary;
  }
  
  try {
    stdLibrary = await parseStandardLibrary(connection, workspaceFolders);
    stdLibraryInitialized = true;
    
    // Log detailed information about what was loaded
    if (stdLibrary && stdLibrary.modules && stdLibrary.functions) {
      const moduleCount = stdLibrary.modules.size;
      const functionCount = stdLibrary.functions.size;
      if (moduleCount > 0 && functionCount > 0) {
        connection.console.log(
          `Standard library initialized: ${moduleCount} modules, ${functionCount} functions available for autocomplete`
        );
      } else {
        connection.console.warn(
          'Standard library initialized but no modules/functions found - autocomplete may not work'
        );
      }
    }
  } catch (error) {
    connection.console.error(`Failed to parse standard library: ${error}`);
    stdLibrary = { modules: new Map(), functions: new Map() };
  }
  
  return stdLibrary;
}

// When the server starts, initialize
connection.onInitialize(async (params) => {
  // Get workspace folders from params
  const workspaceFolders = params.workspaceFolders || null;
  
  // Use workspace root from initialization options if available (fallback)
  const workspaceRoot = (params.initializationOptions as any)?.workspaceRoot;
  if (workspaceRoot && !workspaceFolders) {
    const path = require('path');
    process.env.OMNI_STD_PATH = path.join(workspaceRoot, 'omni/std');
  }
  
  await initializeStdLibrary(workspaceFolders);
  
  const result: InitializeResult = {
    capabilities: {
      textDocumentSync: TextDocumentSyncKind.Incremental,
      completionProvider: {
        triggerCharacters: ['.', ':', '('],
        resolveProvider: true,
      },
      hoverProvider: true,
      signatureHelpProvider: {
        triggerCharacters: ['('],
      },
      definitionProvider: true,
      referencesProvider: true,
      documentSymbolProvider: true,
      workspaceSymbolProvider: {
        resolveProvider: false,
      },
      documentFormattingProvider: true,
      codeActionProvider: {
        codeActionKinds: ['quickfix', 'refactor'],
      },
    },
  };
  
  return result;
});

connection.onInitialized(() => {
  connection.console.log('OmniLang Language Server initialized');
});

// Handle document changes
documents.onDidChangeContent((change: TextDocumentChangeEvent<TextDocument>) => {
  // Parse document and update symbol table
  const uri = change.document.uri;
  try {
    const symbols = parseDocument(change.document.getText());
    symbolTables.set(uri, symbols);
  } catch (error) {
    connection.console.error(`Failed to parse document ${uri}: ${error}`);
  }
  
  // Trigger diagnostics
  handleDiagnostics(change.document, connection, stdLibrary);
});

documents.onDidOpen((event) => {
  const uri = event.document.uri;
  try {
    const symbols = parseDocument(event.document.getText());
    symbolTables.set(uri, symbols);
  } catch (error) {
    connection.console.error(`Failed to parse document ${uri}: ${error}`);
  }
  
  handleDiagnostics(event.document, connection, stdLibrary);
});

documents.onDidClose((event) => {
  symbolTables.delete(event.document.uri);
});

// Register LSP handlers
connection.onCompletion((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  const symbols = symbolTables.get(params.textDocument.uri) || new SymbolTable();
  return handleCompletion(params, document, stdLibrary, symbols);
});

connection.onCompletionResolve((item) => {
  return item;
});

connection.onHover((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  const symbols = symbolTables.get(params.textDocument.uri) || new SymbolTable();
  return handleHover(params, document, stdLibrary, symbols);
});

connection.onSignatureHelp((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  const symbols = symbolTables.get(params.textDocument.uri) || new SymbolTable();
  return handleSignatureHelp(params, document, stdLibrary, symbols);
});

connection.onDefinition((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  const symbols = symbolTables.get(params.textDocument.uri) || new SymbolTable();
  return handleDefinition(params, document, symbols);
});

connection.onReferences((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  const symbols = symbolTables.get(params.textDocument.uri) || new SymbolTable();
  return handleReferences(params, document, symbols, documents);
});

connection.onDocumentSymbol((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  const symbols = symbolTables.get(params.textDocument.uri) || new SymbolTable();
  return handleDocumentSymbol(params, document, symbols);
});

connection.onWorkspaceSymbol((params) => {
  return handleWorkspaceSymbol(params, symbolTables, documents);
});

connection.onDocumentFormatting((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  return handleFormatting(params, document);
});

connection.onCodeAction((params) => {
  const document = documents.get(params.textDocument.uri);
  if (!document) return null;
  
  return handleCodeAction(params, document);
});

// Make the text document manager listen on the connection
documents.listen(connection);

// Listen on the connection
connection.listen();

