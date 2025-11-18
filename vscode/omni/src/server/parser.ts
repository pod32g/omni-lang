export class SymbolTable {
  functions: Map<string, FunctionSymbol> = new Map();
  variables: Map<string, VariableSymbol> = new Map();
  structs: Map<string, StructSymbol> = new Map();
  enums: Map<string, EnumSymbol> = new Map();
  types: Map<string, TypeSymbol> = new Map();
}

export interface FunctionSymbol {
  name: string;
  line: number;
  column: number;
  parameters: Array<{ name: string; type: string }>;
  returnType: string;
  isAsync?: boolean;
}

export interface VariableSymbol {
  name: string;
  line: number;
  column: number;
  type: string;
}

export interface StructSymbol {
  name: string;
  line: number;
  column: number;
  fields: Array<{ name: string; type: string }>;
}

export interface EnumSymbol {
  name: string;
  line: number;
  column: number;
  variants: string[];
}

export interface TypeSymbol {
  name: string;
  line: number;
  column: number;
  aliasedType: string;
}

// Parse OmniLang source code to extract symbols
export function parseDocument(text: string): SymbolTable {
  const symbols = new SymbolTable();
  const lines = text.split('\n');

  for (let lineNum = 0; lineNum < lines.length; lineNum++) {
    const line = lines[lineNum];
    const trimmed = line.trim();

    // Parse function declarations
    // Match: async? func functionName(params):returnType
    const funcMatch = trimmed.match(/(async\s+)?func\s+(\w+)\s*\(([^)]*)\)\s*:?\s*([^{]*?)\s*\{/);
    if (funcMatch) {
      const isAsync = !!funcMatch[1];
      const name = funcMatch[2];
      const paramsStr = funcMatch[3].trim();
      const returnType = (funcMatch[4] || 'void').trim();

      const parameters: Array<{ name: string; type: string }> = [];
      if (paramsStr) {
        const paramParts = paramsStr.split(',').map(p => p.trim());
        for (const param of paramParts) {
          if (!param) continue;
          const paramMatch = param.match(/^(\w+)\s*:\s*(.+)$/);
          if (paramMatch) {
            parameters.push({
              name: paramMatch[1],
              type: paramMatch[2].trim(),
            });
          }
        }
      }

      symbols.functions.set(name, {
        name,
        line: lineNum,
        column: line.indexOf('func'),
        parameters,
        returnType,
        isAsync,
      });
    }

    // Parse variable declarations
    // Match: let|var variableName:type = ...
    const varMatch = trimmed.match(/(let|var)\s+(\w+)\s*:\s*([^=]+?)(\s*=|$)/);
    if (varMatch) {
      const name = varMatch[2];
      const type = varMatch[3].trim();
      symbols.variables.set(name, {
        name,
        line: lineNum,
        column: line.indexOf(name),
        type,
      });
    }

    // Parse struct declarations
    // Match: struct StructName { ... }
    const structMatch = trimmed.match(/struct\s+(\w+)\s*\{/);
    if (structMatch) {
      const name = structMatch[1];
      const fields: Array<{ name: string; type: string }> = [];
      
      // Try to extract fields from the same line or following lines
      let braceCount = (line.match(/\{/g) || []).length;
      let currentLine = lineNum;
      
      while (braceCount > 0 && currentLine < lines.length) {
        const currentLineText = lines[currentLine];
        const fieldMatch = currentLineText.match(/(\w+)\s*:\s*([^,}]+)/g);
        if (fieldMatch) {
          for (const field of fieldMatch) {
            const parts = field.split(':').map(p => p.trim());
            if (parts.length === 2) {
              fields.push({
                name: parts[0],
                type: parts[1],
              });
            }
          }
        }
        braceCount += (currentLineText.match(/\{/g) || []).length - (currentLineText.match(/\}/g) || []).length;
        currentLine++;
      }

      symbols.structs.set(name, {
        name,
        line: lineNum,
        column: line.indexOf('struct'),
        fields,
      });
    }

    // Parse enum declarations
    // Match: enum EnumName { ... }
    const enumMatch = trimmed.match(/enum\s+(\w+)\s*\{/);
    if (enumMatch) {
      const name = enumMatch[1];
      const variants: string[] = [];
      
      // Extract enum variants
      let braceCount = (line.match(/\{/g) || []).length;
      let currentLine = lineNum;
      
      while (braceCount > 0 && currentLine < lines.length) {
        const currentLineText = lines[currentLine];
        const variantMatch = currentLineText.match(/(\w+)(\s*,|\s*\})/g);
        if (variantMatch) {
          for (const variant of variantMatch) {
            const variantName = variant.replace(/[,\s}]/g, '').trim();
            if (variantName) {
              variants.push(variantName);
            }
          }
        }
        braceCount += (currentLineText.match(/\{/g) || []).length - (currentLineText.match(/\}/g) || []).length;
        currentLine++;
      }

      symbols.enums.set(name, {
        name,
        line: lineNum,
        column: line.indexOf('enum'),
        variants,
      });
    }

    // Parse type aliases
    // Match: type TypeName = ...
    const typeMatch = trimmed.match(/type\s+(\w+)\s*=\s*(.+)/);
    if (typeMatch) {
      const name = typeMatch[1];
      const aliasedType = typeMatch[2].trim();
      symbols.types.set(name, {
        name,
        line: lineNum,
        column: line.indexOf('type'),
        aliasedType,
      });
    }
  }

  return symbols;
}

