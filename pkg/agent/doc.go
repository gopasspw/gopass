// Package agent contains a long running background process to aide
// gopass in caching credentials. Since gopass is a one-off application
// it can not store much state in memory. It's lost once gopass quits.
// However certain operations require frequently entering credentials,
// like passphrases for custom crypto backend or encrypted config stores.
// This package implements an agent, similar to the GPG or SSH agents,
// that can ask for and cache credentials.
package agent
