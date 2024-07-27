---
title: a-result-type-for-typescript
published: 2022-12-06
revision: 2022-12-15
tags: ts
excerpt: a Result type is an abstraction to communicate the outcome of fallible operations. In the JS world, we are more used to throwing errors, but this approach has its advantages, especially in message-based communication.
---

When dealing with fallible operations (things that can produce an error), it is normal to _throw_ the error and _catch_ it somewhere else. Or, if the fallible operation is async, reject a _promise_, (or use the "error-first" callback pattern). This has the advantage of being a well-known pattern, so anybody new to the project will quickly understand it.
However, there are situations where these approaches get messy, or are not possible at all.

I was recently working on an Electron app and a browser extension. In both cases, there are two processes that communicate with each other sending messages. Throwing an error is not a great option, because the message sender is not waiting for the response so as to catch any potential errors.

Let's say the main process sends a message to the background process asking it to fetch some information from the database, let's call it `fetch`. The happy path is that the background process successfully retrieves such information and sends a message with the data to the main process, let's call it `data`. So, the main process dispatches a `fetch` event and creates a listener for the `data` event.

But what about errors? Many things can go wrong in the background process.
One idea would be to send `null` or `undefined` as the `data` message, but we loose the ability to communicate information about the error to the main process.

An alternative would be to send an `error` message, instead of a `data` one. This implies that the main process needs two listeners, one for `data` and one for `error`. And what do we do if we have many of this kind of transactions between main and background? Should we have a different error channel for each, or a single one to communicate all errors?

Both ways would work just fine. But I didn't like the idea of having channels dedicated to errors. I wanted to communicate errors or values the same way, and to do it in a TS-friendly way, so I can have type information available without any casting or type guarding.

That's when `Result` comes in. The implementation itself is quite simple, it's just this:

```typescript
type Result<E, V> = { status: "error"; error: E } | { status: "success"; value: V }
```

The _status_ flag and the _union_ allow TS to infer which one of those two types we are working with based on the value of `status` (it's a _discriminated union_).

You may be wondering why not make `Result` generic over the `value` only, instead of both the `error` and the `value`. So, instead of `Result<E, V>`, just `Result<V>`, and then `{ status: "error", error: Error }`.
That would work too, especially if you are not planning on using custom errors, but I am.

Let's go through an example. Imagine we have a voting platform, where we have several proposals and users can vote _Yes_ or _No_. We will create a `Ballot` type like so:

```typescript
type Ballot = {
  proposalId: number
  vote: "Y" | "N"
}
```

We receive ballots and we need to parse them so that we can tally the votes, we need a `parseBallot` function. Now, this is a fallible operation, in this example we will receive the ballot as a JSON string, which means we can have an error if the input is not valid JSON. We can also have an error if it's valid JSON, but not a valid ballot. So, this is a perfect opportunity for us to use a `Result`.
Instead of `parseBallot` returning a `Ballot` or throwing an error, we will return a `Result`.

A small digression, one thing I don't like about throwing errors in general, is that that information is not discoverable via the type system, you have to look at the implementation to learn that something can actually throw an error (or documentation, when I write things that throw, I try to say so in a jsdoc-style comment, since that comment will pop up when using the function, but still not great).

Take this example:

```typescript
function parse(): Ballot {
  throw new Error()
}

const b = parse()
```

Imagine that the `parse` function is defined somewhere else, even in an external library, by just looking at `const b = parse()`, you get that `b` is of type `Ballot` and that the signature of `parse()` is `() => Ballot`, no indication whatsoever that it can throw an error.

End of digression, let's get back to the `parseBallot` function. The first thing we'll do is to create some errors to convey the specific nature of what went wrong, to keep it simple, we'll create only two, but you can imagine having several different errors, especially if we were dealing with HTTP requests and databases to get and store the ballots.

```typescript
class InvalidJSON extends Error {
  constructor() {
    super("payload is not valid JSON.")
    this.name = "InvalidJSON"
  }
}

class InvalidBallot extends Error {
  constructor() {
    super("malformed ballot payload.")
    this.name = "InvalidBallot"
  }
}

// This is not necessary, but it's sometimes useful to have this sort of "grouping".
type ParseError = InvalidJSON | InvalidBallot
```

Now that we have error classes for the two things that we are going to check during parsing, let's go ahead with a minimal version of the parsing function, without extensive validation to keep it focused on what's important here:

```typescript
function parseBallot(jsonString: string): Result<ParseError, Ballot> {
  try {
    const ballot = JSON.parse(jsonString)
    if ("proposalId" in ballot && "vote" in ballot) {
      return { status: "success", value: ballot }
    }
    return { status: "error", error: new InvalidBallot() }
  } catch {
    return { status: "error", error: new InvalidJSON() }
  }
}
```

With this in place, we can use it like so:

```typescript
const ballotResult = parseBallot('{"proposalId": 42, "vote": "Y"}')
if (ballotResult.status === "success") {
  // In this block, TS correctly infers that we're dealing with:
  //    `{ status: "success"; value: Ballot }`
  const count = ballotResult.value.vote === "Y" ? "+1" : "-1"
  console.log(`${count} for event ${ballotResult.value.proposalId}`)
  // ballotResult.error // -> property `error` does not exist...
} else {
  // In this block, TS correctly infers that we're dealing with:
  //    `{ status: "error"; error: ParseError }`
  console.error(`${ballotResult.error.name}: ${ballotResult.error.message}`)
  // ballotResult.value // -> property `value` does not exist...
}
```

Try changing the input of the `parseBallot` function above, to malformed JSON and to valid JSON but not a valid `Ballot`.

It will also infer the correct types using ternary operator:

```typescript
const _a = ballotResult.status === "success"
  // type here is { status: "success"; value: Ballot }
  ? ballotResult.value.proposalId
  // type here is { status: "error"; error: ParseError }
  : ballotResult.error.name
```

Node is using this exact same approach for its `PromiseSettledResult`, it looks something like:

```typescript
interface PromiseFulfilledResult<T> {
  status: "fulfilled";
  value: T;
}

interface PromiseRejectedResult {
  status: "rejected";
  reason: any;
}

type PromiseSettledResult<T> = PromiseFulfilledResult<T> | PromiseRejectedResult;
```

.
