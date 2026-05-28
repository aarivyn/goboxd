 # Languages

All languages are registered in config/languages.yaml. Adding a new language requires only a YAML block and no Go code change.

## Supported

| ID | Name | Type |
|----|------|------|
| py3 | Python 3 | interpreted |
| cpp | C++ | compiled |
| c | C | compiled |
| java | Java | compiled |
| bash | Bash | interpreted |
| js | JavaScript (Node) | interpreted |
| verilog | Verilog | compiled |

## Adding a language

1. Install the compiler or runtime in the Dockerfile
2. Add a block to config/languages.yaml with id, name, source_filename, build (if compiled), and run
3. Rebuild the Docker image

No Go code change needed.
