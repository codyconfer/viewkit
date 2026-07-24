// Package panels renders charts, tables, and small status widgets against a
// viewkit layout frame.
//
// DualHost (host.go) is the inline-shell vs deck mounting contract: panels stay
// tea-free and register via Register/Lookup for plugin contribution.
package panels
