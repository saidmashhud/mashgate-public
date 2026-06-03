module github.com/saidmashhud/mashgate-public/examples/go

go 1.22

require github.com/saidmashhud/mashgate-public/sdk/go v0.0.0

// Build against the SDK source in this repo instead of fetching it from GitHub.
// Remove this block (and run `go get github.com/saidmashhud/mashgate-public/sdk/go@main`)
// when copying this example into your own project.
replace github.com/saidmashhud/mashgate-public/sdk/go => ../../sdk/go

require github.com/google/uuid v1.6.0 // indirect
