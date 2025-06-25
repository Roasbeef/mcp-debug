package mcpdebug

// LaunchConfig represents the configuration for launching a Go program
// for debugging. This provides a type-safe wrapper around the DAP launch
// arguments.
type LaunchConfig struct {
	// Name is a human-readable name for the debug session.
	Name string

	// Program specifies the path to the Go program to debug. This can be
	// either a path to a Go source file or a directory containing a main
	// package.
	Program string

	// Args contains the command-line arguments to pass to the program
	// being debugged.
	Args []string

	// Env contains environment variables to set for the debugged program.
	// Each entry should be in the format "KEY=value".
	Env []string

	// WorkingDir specifies the working directory for the debugged program.
	// If empty, the current directory is used.
	WorkingDir string

	// StopOnEntry determines whether to stop at the program's entry point.
	StopOnEntry bool

	// BuildFlags contains additional flags to pass to the Go compiler
	// when building the program for debugging.
	BuildFlags []string
}

// AttachConfig represents the configuration for attaching to an existing
// running process for debugging.
type AttachConfig struct {
	// Name is a human-readable name for the debug session.
	Name string

	// ProcessID is the ID of the process to attach to.
	ProcessID int

	// Mode specifies the attach mode. Valid values are "local" for local
	// processes and "remote" for remote debugging.
	Mode string

	// Host and Port are used for remote debugging when Mode is "remote".
	Host string
	Port int
}

// BreakpointLocation represents a location where a breakpoint can be set.
type BreakpointLocation struct {
	// File is the path to the source file.
	File string

	// Line is the line number (1-based) where the breakpoint should be set.
	Line int

	// Column is the column number (1-based) where the breakpoint should be
	// set. This is optional and can be 0 for line-based breakpoints.
	Column int

	// Condition is an optional condition that must be true for the
	// breakpoint to be hit.
	Condition string

	// HitCondition specifies when the breakpoint should be hit based on
	// hit count (e.g., ">= 5" to break after 5 hits).
	HitCondition string

	// LogMessage specifies a message to log when the breakpoint is hit
	// instead of stopping execution.
	LogMessage string
}

// FunctionBreakpoint represents a breakpoint set on a function name.
type FunctionBreakpoint struct {
	// Name is the name of the function to break on.
	Name string

	// Condition is an optional condition that must be true for the
	// breakpoint to be hit.
	Condition string

	// HitCondition specifies when the breakpoint should be hit based on
	// hit count.
	HitCondition string
}

// ThreadInfo represents information about a thread in the debugged program.
type ThreadInfo struct {
	// ID is the unique identifier for the thread.
	ID int

	// Name is the human-readable name of the thread.
	Name string
}

// StackFrame represents a single frame in the call stack.
type StackFrame struct {
	// ID is the unique identifier for this frame.
	ID int

	// Name is the name of the function for this frame.
	Name string

	// Source contains information about the source file for this frame.
	Source SourceInfo

	// Line is the line number in the source file (1-based).
	Line int

	// Column is the column number in the source file (1-based).
	Column int
}

// SourceInfo represents information about a source file.
type SourceInfo struct {
	// Path is the full path to the source file.
	Path string

	// Name is the name of the source file without path.
	Name string
}

// VariableScope represents a scope containing variables (e.g., local, global).
type VariableScope struct {
	// Name is the human-readable name of the scope.
	Name string

	// VariablesReference is the reference ID used to retrieve variables
	// in this scope.
	VariablesReference int

	// Expensive indicates whether retrieving variables from this scope
	// is expensive and should be done lazily.
	Expensive bool
}

// Variable represents a variable and its value in the debugged program.
type Variable struct {
	// Name is the name of the variable.
	Name string

	// Value is the string representation of the variable's value.
	Value string

	// Type is the type of the variable.
	Type string

	// VariablesReference is the reference ID for child variables if this
	// variable is a complex type (struct, array, etc.). A value of 0
	// indicates no child variables.
	VariablesReference int

	// IndexedVariables is the number of indexed child variables if this
	// variable is an array or slice.
	IndexedVariables int

	// NamedVariables is the number of named child variables if this
	// variable is a struct or map.
	NamedVariables int
}

// EvaluationResult represents the result of evaluating an expression.
type EvaluationResult struct {
	// Result is the string representation of the evaluation result.
	Result string

	// Type is the type of the result.
	Type string

	// VariablesReference is the reference ID for child variables if the
	// result is a complex type.
	VariablesReference int

	// IndexedVariables is the number of indexed child variables.
	IndexedVariables int

	// NamedVariables is the number of named child variables.
	NamedVariables int
}