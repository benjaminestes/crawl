// Package data provides types appropriate for describing the output
// of a crawler. All of the fields of these types are exported, since
// they are intended to be marshalled into some transmission format.
//
// In general, the approach is to define simple concepts and embed
// them in more complex types when that will simplify
// implementation. For instance, an Address object describes a single
// URL. A Link object describes a link scraped from a webpage.  this
// Link has an Address to which it points, but it also has anchor text
// which might be interesting to analyze.
package data
