// Package plans holds the sources that go run …/cmd/portico generate copies into a site.
// Each file is one installation: it sets a root and Build succeeds alone.
// Sites do not import this package. They edit the generated copy.
//
// The "How to change this plan" comment at the top of each file is the
// guide for that copy. Keep it accurate when the commands or effects change.
// guestbook and chat share rest.go; generate writes it with either of them.
package plans
