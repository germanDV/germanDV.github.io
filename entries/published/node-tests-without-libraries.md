---
title: node-tests-without-libraries
published: 2022-12-29
revision: 2022-12-29
excerpt: Let's test some Javascript code without using any libraries, just a couple of JS files, no package.json, no npm.
---

We will be using the new _test runner_ module from Node, so make sure to have a current version installed. I'm on the latest LTS at the time of this writing: **18.12.1**. The _test runner_ module is experimental (stability: 1). Which means:

> Stability: 1 - Experimental. The feature is not subject to semantic versioning rules. Non-backward compatible changes or removal may occur in any future release. Use of the feature is not recommended in production environments.

Therefore, you are probably better off sticking to Jest, or whatever testing library you are using, for production, for now.

Because we won't be using any libraries, and we won't be using npm (nor npx) at all, we won't be using Typescript either, we want our script and tests to be run just with `node`.

## The Feature

In order not to write tests that assert that `2 + 2 === 4`, let's create a more or less useful function to test. We will create a `parseRoute` function that given a template and an actual route will extract variables define in the template from the route. So, for example, given the template `/user/:id` and the route `/user/10`, `parseRoute` should return something like `{ id: 10 }`. And just for fun, let's also parse the query string, if it exists.

We cannot use Typescript, but that doesn't mean we cannot type our program, we will be using [JSDoc](https://jsdoc.app/) for that.

Let's get started. Create a file called `index.mjs`, I'm using the _.mjs_ extension so that we can use **ES Modules** for our imports and exports.

```javascript
// index.mjs

/**
 * An object containing the variables in a path,
 * including the query string params, if any
 * @typedef {Object} Path
 * @property {string} route The original route
 * @property {Object<string, string>} params The dynamic params from the template
 * @property {Object<string, string>} query The query string params
 */

/**
 * Parses a route using a template that can include variables.
 * Variables start with `:`, for example `/user/:id`.
 * If the route does not match the template, it returns null.
 *
 * Each variable in the template becomes a key in a `params` object
 * in the returned Path object.
 *
 * The query string is also parsed and key values are stored in a `query` key
 * in the returned Path object.
 *
 * @param {string} template - The template to match the route against.
 * @param {string} route - The actual route to parse.
 * @returns {Path|null} The parsed route.
 */
export function parseRoute(template, route) {
  if (!template || !route) {
    throw new Error("template and route params are required")
  }

  /** @type {Object<string, string>} */
  const query = {}

  // Check if route includes a query string.
  if (route.includes("?")) {
    const qs = route.substring(route.indexOf("?"))
    route = route.substring(0, route.indexOf("?"))
    qs.replace(
      new RegExp("([^?=&]+)(=([^&]*))?", "g"),
      (_, k, __, v) => (query[k] = v)
    )
  }

  const templateParts = template.split("/")
  const routeParts = route.split("/")
  if (templateParts.length !== routeParts.length) return null

  /** @type Path */
  const path = {
    route, // add the original route to the `path` for convenience.
    query, // add the parsed query string.
    params: {}, // an empty object to collect variables from the template.
  }

  for (let i = 0; i < templateParts.length; i++) {
    // This part of the template is a variable, add it to the `path`
    // object to be returned.
    if (templateParts[i].startsWith(":")) {
      path.params[templateParts[i].substring(1)] = routeParts[i]
      continue
    }

    // If template part is not a variable and it does not match the
    // corresponding route part, then the route does not match the
    // template and we will return `null`.
    if (!templateParts[i].startsWith(":") && templateParts[i] !== routeParts[i]) {
      return null
    }
  }

  return path
}

/**
 * Just to show how to test async functions,
 * let's have a promisified version of `parseRoute`.
 *
 * @param {string} template - The template to match the route against.
 * @param {string} route - The actual route to parse.
 * @returns {Promise<Path|null>} The parsed route.
 */
export  function parseRouteAsync(template, route) {
  return new Promise((resolve, reject) => {
    try {
      const path = parseRoute(template, route)
      if (!path) {
        reject(new Error("route could not be matched to template."))
      } else {
        resolve(path)
      }
    } catch (err) {
      reject(err)
    }
  })
}
```

Some things worth noticing:

* We have defined our own `Path` type at the top of the file.
* We use the comment `/** @type {Object<string, string>} */` to annotate the type of a variable.
* In some cases the function throws an error and in others it returns `null`, this is so that we have more things to test.
* If you want to enable type checks for the file on your text editor, you can add a `//@ts-check` comment at the top of the file.
* Another way to type check the file would be running `tsc --allowJs --checkJs --noEmit --target ESNext index.mjs` (assuming you have typescript installed on your system).
* If you think `parseRouteAsync` makes no sense, that's right, it's just there so that we can test promises.

<br/>
## The Tests

Let's now create a test file, I like calling test files `[whatever].test.mjs`, and putting them next to the file they are testing. But there are many approaches in terms of naming and placement, so you do you. My only advice is to follow one of the patterns that node is aware of, so that it can pick up test files automatically. [Check the test runner execution model here](https://nodejs.org/dist/latest-v18.x/docs/api/test.html#test-runner-execution-model).

```javascript
// index.test.mjs

import { describe, it, test } from "node:test"
import assert from "node:assert/strict"
import { parseRoute, parseRouteAsync } from "./index.mjs"

describe("parseRoute", () => {
  it("should throw when template is empty", () => {
    assert.throws(() => parseRoute("", "/post/123"), Error)
  })

  it("should throw when route is empty", () => {
    assert.throws(() => parseRoute("/user/:id", ""), Error)
  })

  it("should return null when route does not match template", () => {
    const got = parseRoute("/user/:id", "/post/123")
    assert.equal(got, null)
  })

  it("should correctly parse a route without variables in the template", () => {
    const want = { route: "/about/contact", params: {}, query: {} }
    const got = parseRoute("/about/contact", "/about/contact")
    assert.deepEqual(got, want)
  })

  it("should return null if trailing slashes do not match in route and template", () => {
    const got = parseRoute("/info", "/info/")
    assert.equal(got, null)
  })

  it("should collect variables according to template", () => {
    const want = {
      route: "/user/99/post/458",
      params: { userId: "99", postId: "458" },
      query: {},
    }

    const got = parseRoute("/user/:userId/post/:postId", "/user/99/post/458")
    assert.deepEqual(got, want)
  })

  it("should parse variables and query string", () => {
    const want = {
      route: "/user/99",
      params: { id: "99" },
      query: { lang: "it" },
    }

    const got = parseRoute("/user/:id", "/user/99?lang=it")
    assert.deepEqual(got, want)
  })

  it("should parse multiple variables and query string params", () => {
    const want = {
      route: "/user/99/post/458",
      params: { userId: "99", postId: "458" },
      query: { lang: "it", format: "json" },
    }

    const got = parseRoute(
      "/user/:userId/post/:postId",
      "/user/99/post/458?lang=it&format=json",
    )

    assert.deepEqual(got, want)
  })
})
```

Some things worth noticing:

* We `import assert from "node:assert/strict"` so that we always use the strict version and can write `assert.equal` instead of `assert.strictEqual`.
* We could use the `test` that we import from `node:test` to replace all the `describe`s and `it`s. But for familiarity with other testing libraries, I'm using `describe` and `it` for this first part, we'll use `test` later.
* I haven't found a good use case for the `before`, `beforeEach`, `after` and `afterEach` hooks here, but they are available, check the [docs](https://nodejs.org/dist/latest-v18.x/docs/api/test.html) for further details.

To run the tests, we can simply run `node --test`, the results are reported in [TAP](https://testanything.org/) format. When you run it, you should see something like:

```sh
AP version 13
# Subtest: index.test.mjs
ok 1 - index.test.mjs
  ---
  duration_ms: 49.6767
  ...
1..1
# tests 1
# pass 1
# fail 0
# cancelled 0
# skipped 0
# todo 0
# duration_ms 53.215802
```

Make one of the tests fail and check the report.


## Testing Async

Let's add another subtest to test our `parseRouteAsync` function. The [assert](https://nodejs.org/dist/latest-v18.x/docs/api/assert.html) module has utilities to test Promise rejections; however, I could not make it work. Maybe that's because the test runner is still experimental and doesn't handle async functions well yet, or, more likely, I'm an idiot making some silly mistake.

In any case, I thought it would be a good idea to show a workaround, because sometimes a workaround is the best we can do, at least temporarily, when time is of the essence. Hence, we will create a small helper function to help us test if a promise rejects.

```
// index.test.mjs

...

describe("parseRoute", () => {
  ...
})

test("parseRouteAsync", async (t) => {
  await t.test("should collect variables according to template", async () => {
    const want = {
      route: "/user/77/post/919",
      params: { userId: "77", postId: "919" },
      query: {},
    }

    const got = await parseRouteAsync("/user/:userId/post/:postId", "/user/77/post/919")
    assert.deepEqual(got, want)
  })

  await t.test("should reject when inputs are empty", async () => {
    const rejected = await rejects(parseRouteAsync("", ""))
    assert.ok(rejected, "promise did not reject")
  })
})

/**
 * Helper function to check if a promise rejects with an Error.
 * @param {Promise} promise - The promise to check.
 * @returns {Promise<boolean>} Whether the promise rejected or not.
 */
async function rejects(promise) {
    let rejectsWithError = false
    try {
      await promise
    } catch (err) {
      if (err instanceof Error) {
        rejectsWithError = true
      }
    }
    return rejectsWithError
}
```

Some things worth noticing:

* We are using `test` instead of `describe` and `it`. When using `test`, the callback receives a `t` argument that we can use to call subtests, call hooks like `beforeEach`, and more. So, generally speaking, I would prefer `test` over `describe` and `it`.
* When using `test`, it is important to **always `await` subtests, even if they are not async**, from the docs: "_This is necessary because parent tests do not wait for their subtests to complete. Any subtests that are still outstanding when their parent finishes are cancelled and treated as failures._".

<br />
## Conclusion

Other than the `assert.rejects` fiasco (which is more than likely on me), I like that node provides a native test runner. But remember that it's unstable and not production-ready yet.
