# Blog Engine

- `make help` -> print help message about commands in Makefile.
- `make dev` -> start development server. Preview drafts by going to `/preview/` in the browser.
- `make build` -> build binary called `gdv` (place it in PATH or make an alias).
- `gdv -h` -> print help message about blog commands.
- `gdv -draft title-of-new-entry` -> create markdown layout of a new entry in the _drafts_ folder.
- `gdv -serve` -> start web server.
- `gdv -publish` -> provide a list of drafts, choose which one to publish.
- `gdv -publish-all` -> publish all drafts.
- `gdv -feed` -> generate/update the RSS feed. Most of the times, you'll want to run this after publishing.
