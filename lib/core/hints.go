package core

var hints = map[string]string{
	"file_not_found":       "use 'zipts .' to create a new archive from the current directory",
	"missing_arg":          "run the command with '--help' to see usage and examples",
	"invalid_zip":          "verify the file is a valid ZIP archive with 'file <name>'",
	"entry_not_found":      "use 'zipls <archive>' to list available entries",
	"output_dir_not_found": "create the directory first with 'mkdir -p <path>'",
	"nested_zip_not_found": "check the nesting path with 'zipls <outer.zip>'",
	"params_file_missing":  "create a zipw.yml file or pass params via -f, -z and -i flags",
}

// Hint returns a user-facing suggestion string for the given error context key.
func Hint(key string) string {
	return hints[key]
}
