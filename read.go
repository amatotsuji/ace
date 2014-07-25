package ace

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Special characters
const (
	cr   = "\r"
	lf   = "\n"
	crlf = "\r\n"

	space        = " "
	equal        = "="
	pipe         = "|"
	doublePipe   = pipe + pipe
	slash        = "/"
	sharp        = "#"
	dot          = "."
	doubleDot    = dot + dot
	colon        = ":"
	doubleQuote  = `"`
	lt           = "<"
	gt           = ">"
	exclamation  = "!"
	hyphen       = "-"
	bracketOpen  = "["
	bracketClose = "]"
)

// readFiles reads files and returns source for the parsing process.
func readFiles(bathPath, innerPath string, opts *Options) (*source, error) {
	// Read the base file.
	base, err := readFile(bathPath, opts)
	if err != nil {
		return nil, err
	}

	// Read the inner file.
	inner, err := readFile(innerPath, opts)
	if err != nil {
		return nil, err
	}

	var includes []*file

	// Find include files from the base file.
	if err := findIncludes(base.data, opts, &includes); err != nil {
		return nil, err
	}

	// Find include files from the inner file.
	if err := findIncludes(inner.data, opts, &includes); err != nil {
		return nil, err
	}

	return newSource(base, inner, includes), nil
}

// readFile reads a file and returns a file struct.
func readFile(path string, opts *Options) (*file, error) {
	data, err := ioutil.ReadFile(path + "." + opts.Extension)
	if err != nil {
		return nil, err
	}

	return NewFile(path, data), nil
}

// findIncludes finds and adds include files.
func findIncludes(data []byte, opts *Options, includes *[]*file) error {
	includePaths, err := findIncludePaths(data, opts)
	if err != nil {
		return err
	}

	for _, includePath := range includePaths {
		if !hasFile(*includes, includePath) {
			f, err := readFile(includePath, opts)
			if err != nil {
				return err
			}

			*includes = append(*includes, f)

			if err := findIncludes(f.data, opts, includes); err != nil {
				return err
			}
		}
	}

	return nil
}

// findIncludePaths finds and returns include paths.
func findIncludePaths(data []byte, opts *Options) ([]string, error) {
	var includePaths []string

	for i, str := range strings.Split(formatLF(string(data)), lf) {
		ln := newLine(i+1, str, opts)

		if ln.isHelperMethodOf(helperMethodNameInclude) {
			if len(ln.tokens) < 3 {
				return nil, fmt.Errorf("no template name is specified [line: %d]", ln.no)
			}

			includePaths = append(includePaths, ln.tokens[2])
		}
	}

	return includePaths, nil
}

// formatLF replaces the line feed codes with LF and returns the result.
func formatLF(s string) string {
	return strings.Replace(strings.Replace(s, crlf, lf, -1), cr, lf, -1)
}

// hasFile return if files has a file which has the path specified by the parameter.
func hasFile(files []*file, path string) bool {
	for _, f := range files {
		if f.path == path {
			return true
		}
	}

	return false
}
