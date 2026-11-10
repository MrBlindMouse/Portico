# internal/plans

Plan sources copied into a site by `portico generate`. Not a site import.

Each file returns `*installation.Builder`. One plan, one purpose, and it builds alone (it sets a root). Do not add an action with no scene. `guestbook` and `chat` call `restAfter` in `rest.go`; generate writes that file with them.

Sites edit the copy. Do not add a site's domain logic here.
