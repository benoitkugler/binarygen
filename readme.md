# binarygen, a Golang code generator for Opentype files 

This tool extends [go/packages] to understand the syntax describing
the binary layout used is Opentype font files, and generates Go parsing and writing functions.
 
## Custom syntax 

The binary layout is specified in Go source files using struct tags :

- 'arrayCount' : FirstUint16 | FirstUint32 | ToEnd | ComputedField-<XXX>
- 'offsetSize' : Offset16 | Offset32
- 'subsliceStart' : AtStart | AtCurrent
- 'unionField' : the name of a previous field 
- 'isOpaque' : anything (even the empty string)

The special comment `// binarygen: startOffset=2` indicates that the table starts at `src[2:]`