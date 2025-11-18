import {
  DocumentFormattingParams,
  TextEdit,
  TextDocument,
} from 'vscode-languageserver';

export function handleFormatting(
  params: DocumentFormattingParams,
  document: TextDocument
): TextEdit[] {
  const text = document.getText();
  const lines = text.split('\n');
  const formattedLines: string[] = [];
  
  let indentLevel = 0;
  const indentSize = 4; // Use 4 spaces for indentation

  for (let i = 0; i < lines.length; i++) {
    let line = lines[i];
    const trimmed = line.trim();
    
    // Skip empty lines
    if (trimmed === '') {
      formattedLines.push('');
      continue;
    }

    // Decrease indent before closing braces
    if (trimmed.startsWith('}') || trimmed.startsWith(']') || trimmed.startsWith(')')) {
      indentLevel = Math.max(0, indentLevel - 1);
    }

    // Apply indentation
    const indent = ' '.repeat(indentLevel * indentSize);
    formattedLines.push(indent + trimmed);

    // Increase indent after opening braces
    if (trimmed.endsWith('{') || trimmed.endsWith('[') || trimmed.endsWith('(')) {
      indentLevel++;
    }
  }

  const formattedText = formattedLines.join('\n');
  
  // Only return edit if text changed
  if (formattedText !== text) {
    return [
      TextEdit.replace(
        {
          start: { line: 0, character: 0 },
          end: { line: lines.length - 1, character: lines[lines.length - 1].length },
        },
        formattedText
      ),
    ];
  }

  return [];
}

