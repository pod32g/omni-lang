# OmniLang Grammar (Bootstrap Draft)

The complete OmniLang grammar will be captured using an extended BNF once the
lexer and parser land. For the bootstrap milestone we include the high level
non-terminals to anchor upcoming work.

```
Module      ::= ImportDecl* TopLevelDecl*
ImportDecl  ::= "import" QualifiedName
TopLevelDecl::= StructDecl | EnumDecl | FuncDecl | LetDecl | VarDecl
```

Additional productions will join as language features graduate from design to
implementation. Each change should keep this document authoritative for the Go
parser and supporting tooling.
