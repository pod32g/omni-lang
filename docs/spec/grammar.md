# OmniLang Grammar (MVP)

program        := { declaration }
declaration    := importDecl | varDecl | funcDecl | structDecl | enumDecl

importDecl     := "import" ident { "." ident } [ "as" ident ]
varDecl        := ("let" | "var") ident ":" type "=" expr
funcDecl       := "func" ident "(" params? ")" [ ":" type ] ( block | "=>" expr )
structDecl     := "struct" ident "{" { ident ":" type } "}"
enumDecl       := "enum" ident "{" { ident } "}"

params         := param { "," param }
param          := ident ":" type

type           := primType | arrayType | mapType | ident
primType       := "int" | "long" | "byte" | "float" | "double" | "bool" | "char" | "string" | "void"
arrayType      := "array" "<" type ">"
mapType        := "map" "<" type "," type ">"

block          := "{" { statement } "}"
statement      := varDecl
                | exprStmt
                | ifStmt
                | forStmt
                | block
exprStmt       := expr

ifStmt         := "if" expr block [ "else" ( block | ifStmt ) ]
forStmt        := "for" ( forC | forIn )
forC           := ident ":" "int" "=" expr ";" expr ";" expr block
forIn          := ident "in" expr block

expr           := assignment
assignment     := logic_or [ "=" assignment ]
logic_or       := logic_and { "||" logic_and }
logic_and      := equality { "&&" equality }
equality       := comparison { ( "==" | "!=" ) comparison }
comparison     := term { ( ">" | ">=" | "<" | "<=" ) term }
term           := factor { ( "+" | "-" ) factor }
factor         := unary  { ( "*" | "/" | "%" ) unary }
unary          := ( "!" | "-" | "+" ) unary | call
call           := primary { "(" args? ")" | "." ident }
args           := expr { "," expr }
primary        := literal | ident | "(" expr ")"

literal        := INT | FLOAT | STRING | "true" | "false"
ident          := IDENT
