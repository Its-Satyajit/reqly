package variables

// Variables implements the scoped variable system with explicit precedence.
//
// Scopes, from lowest to highest priority:
//
//	Global → Environment → Collection → Folder → Request → Runtime
//
// A value defined in a higher-priority scope overrides the same key in a
// lower-priority scope.
