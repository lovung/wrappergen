# Wrapper Generator

Wrapper Generator is a tool written in Golang that generates wrapper code for specified interfaces in a given Go package.

## Installation

To install the Wrapper Generator, run the following command:

```bash
go install github.com/lovung/wrappergen
```

## Usage

```bash
wrappergen [flags]
```

### Flags

- `-package_path`: Package path (default is the current working directory).
- `-wrapper_prefix`: Prefix to be added to generated wrapper types.
- `-wrapper_suffix`: Suffix to be added to generated wrapper types (default is "Wrapper").
- `-custom_code`: Custom code to be inserted into generated wrapper methods.
- `-custom_imports`: Custom imports to be added to the generated wrapper file.
- `-exclude_files`: Comma-separated list of file names to ignore during generation.
- `-exclude_interfaces`: Comma-separated list of interface names to ignore during generation.

### Environment Variables

- `CUSTOM_CODE`: Set the custom code using this environment variable.
- `CUSTOM_IMPORTS`: Set the custom imports using this environment variable.

## Examples

Generate wrappers for all interfaces in the current package:

```bash
wrappergen
```

Generate wrappers for interfaces with custom code and imports:

```bash
wrappergen -custom_code="log.Println(\"Custom code here\")" -custom_imports="fmt"
```

Generate wrappers excluding specific files and interfaces:

```bash
wrappergen -exclude_files="file1.go,file2.go" -exclude_interfaces="Interface1,Interface2"
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
