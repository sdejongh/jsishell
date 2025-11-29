package builtins

// RegisterAll registers all built-in commands to the given registry.
func RegisterAll(r *Registry) {
	// Set registry for help command
	SetHelpRegistry(r)

	// Core commands
	r.Register(EchoDefinition())
	r.Register(ExitDefinition())
	r.Register(HelpDefinition())
	r.Register(ClearDefinition())
	r.Register(EnvDefinition())

	// File system commands
	r.Register(CdDefinition())
	r.Register(PwdDefinition())
	r.Register(LsDefinition())
	r.Register(MkdirDefinition())
	r.Register(CpDefinition())
	r.Register(MvDefinition())
	r.Register(RmDefinition())
	r.Register(SearchDefinition())

	// Configuration commands
	r.Register(ReloadDefinition())
	r.Register(InitDefinition())

	// History commands
	r.Register(HistoryDefinition())
}

// RegisterCoreCommands registers only the core commands needed for basic operation.
// Use this for minimal shell functionality.
func RegisterCoreCommands(r *Registry) {
	SetHelpRegistry(r)
	r.Register(EchoDefinition())
	r.Register(ExitDefinition())
	r.Register(HelpDefinition())
}
