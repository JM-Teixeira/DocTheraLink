// Package version holds the Klyniva release number (semantic versioning:
// major.minor.patch; still major 0 while the product takes shape). The same
// number is in fe/package.json - version_test.go checks they agree, and
// scripts/bump-version.sh changes both.
package version

// Version is the current release.
const Version = "0.15.0"
