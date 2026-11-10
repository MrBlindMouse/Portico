# internal/plans

Plan sources copied into a site by `portico generate`. Not a site import.

Each file returns `*installation.Builder`. One plan, one purpose, and it builds alone (it sets a root). Do not add an action with no scene. `guestbook` and `chat` call `restAfter` in `rest.go`; generate writes that file with them.

Each source starts with a **How to change this plan** comment. That block is what site authors and agents read after generate. Update it in the same change as the commands or effects. Do not rely on this AGENTS file for site-facing guidance; it is not copied.

Sites edit the copy. Do not add a site's domain logic here.
