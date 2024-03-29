---
title: hosting-static-site-with-github-pages
published: 2023-03-14
revision: 2023-03-14
excerpt: If you are facing errors trying to deploy a static site generated by anything other than Jekyll to GitHub Pages, a `.nojekyll` file is all you need.
---

I started this site as a Go server, which I deployed to [Render](https://render.com/).
Although this could simply be a static site, actually the Go server just serves precompiled HTML pages from
Markdown entries, I wanted a server to add some more functionality to it.

I haven't totally changed my mind about the extra functionality that would require me to have a server that would
do more than just serving static assets; however, I decided to turn it into a static site that I could host
on GitHub Pages until I developed such functionality, at which point I may come back to Render or use a different
service.

When you host on GitHub Pages, you can either serve the root (`/`) of your repository, or serve files from a `docs/`
directory. In many cases, the latter is the option you want, as the root may contain non-distributable source files
such as the markdown entries, the code to compile to HTML, etc. That's true in my case as I have all the Go code for
what you may call a "blog engine". And it may be true in many other cases; for example, I helped a friend build and
deploy a static site generated with [Next.js](https://nextjs.org), and it's the same thing, you just want the output
directory (which you need to name _docs_) to be served.

In both experiences deploying to GitHub Pages, I ran into the same problem, that's why I decided to write this short article.
After setting everything up in GitHub and pushing to the _main_ branch, the process would fail, with some errors
that I couldn't totally make sense of.

In my case, it was complaining about Go's HTML template syntax, but that was not in the _docs/_ directory,
which gave me the clue that GitHub was not just serving the _docs/_ directory, it was doing some more checks.
After a quick duckduckgo search, I learnt that GitHub treats your site as a [Jekyll](https://jekyllrb.com/) one
by default, which comes with its own rules and expectations.

The solution I found is to simply include a `.nojekyll` empty file in the _docs/_ directory.
And that's it, on the next push, everything builds correctly and gets published as expected.

Hope this saves you some time.
