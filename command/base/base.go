package base

// Command is the base command.
type Command struct {
	DryRun bool `long:"dry-run" description:"Dry run executions of command."`
}
