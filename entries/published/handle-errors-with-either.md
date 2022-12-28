---
title: handle-errors-with-either
published: 2022-12-28
revision: 2022-12-28
excerpt: With an `Either` construct, we can treat errors as values and handle them in an elegant way.
---

In [A Result Type For Typescript](/blog/a-result-type-for-typescript.html), we created a small abstraction so that we could handle errors in a way that does not require the typical throwing-and-catching pattern.

We will take it a step further and have a construct that will also allow us to chain operations, and do so safely even if we run into an error at any point. Let's borrow some functional programming concepts to do it.

We will create an `Either` class that will wrap a value. We will be able to apply functions to that value using a `map()` method, which will return another `Either`, that way, we can keep chaining operations.

`Either`, as its name indicates, can have one of two values, I'm quite tempted to call them _error_ and _value_, but I think these are known as _Left_ and _Right_ in the FP world, so let's stick to the convention. If we run into an error, we will represent it as a _Left_ value; _Right_ will represent the happy path.

When `map`ping over an `Either`, if the value is a _Left_, it will do nothing and simply return the `Either` instance. If it is a _Right_, it will apply the provided function and return the output wrapped in an `Either`.

Let's start writing some code:

```typescript
// either.ts

// I went with the `"Left" | "Right"` intersection, but you could define an Enum if you prefer.
export default class Either<L, R, Type extends "Left" | "Right" = "Right"> {
  // We keep track of the type so it's easy to decide what to do in each scenario.
  private readonly type: "Left" | "Right"

  // If the Type is "Left", the value must be of type L (typically an Error),
  // otherwise it must be of type R.
  private readonly value: Type extends "Left" ? L : R

  // The constructor does not need to be private, we make it so because we have a static `from`
  // method below to do the same job.
  private constructor(value: Type extends "Left" ? L : R, type: "Left" | "Right" = "Right") {
    this.type = type
    this.value = value;
  }

  // We use some default `never`s so that we don't have to provide
  // explicit types when creating Either instances.
  static right<L = never, V = never>(value: V): Either<L, V> {
    return new Either<L, V, "Left" | "Right">(value, "Right");
  };

  static left<V = never, R = never>(value: V): Either<V, R> {
    return new Either<V, R, "Left" | "Right">(value, "Left");
  };

  static from<E = Error, V = any>(value: V): Either<E, V> {
    return new Either<E, V, "Right">(value, "Right")
  }

  private isLeft(): boolean {
    return this.type === "Left"
  }

  // map applyes a function to the underlying value and
  // returns a new Either.
  map<T>(fn: (x: R) => T): Either<L, T> {
    if (this.isLeft()) {
      // the value is a `left` (an error), so we don't
      // run the function.
      return Either.left(this.value as L)
    }
    try {
      // We run the function and wrap the output in an Either.
      // We do it in a try/catch so that the caller does not need
      // to worry about error handling. If we find an error, 
      // we return a `left` instead of a `right`.
      return Either.right(fn(this.value as R))
    } catch (err) {
      return Either.left(err as L)
    }
  }
}
```

With the above code in place, we can create an `Either` and use the `map` method to operate on its value. It will also handle errors for us. Let's do some simple tests:

```typescript
// index.ts
const duplicate = (n: number) => n*2
const flipSign = (n: number) => -n
const square = (n: number) => Math.pow(n, 2)

const n = Either.from(4)
  .map(duplicate)
  .map(square)
  .map(flipSign)

console.log(n)
```

We instantiate Either using the `from` static method. To keep it simple, we do it with a number, but we could have more complex data here, as we'll see later.

We apply three different operations on the number using `map`. At the end, we expect our value to be **-64** (Math.pow(4*2, 2) * -1). When we log it, we see something like:

```sh
Either { type: 'Right', value: -64 }
```

That's good, the `type` is _Right_ because we have not run into any errors, and the `value` is the expected **-64**. We don't have a way of getting the underlying value out, it's a `private` property, we will deal with that soon, but let's see what would happen if we introduced an error.

First, let's introduce a function that could throw an error. We will throw if we try to divide by zero.

```typescript
// Notice that `divider` is a function that takes the divisor and returns a
// function to actually perform the division.
// It provides a more readable approach to me, but it is not strictly required,
// so no worries if you don't like it, you can do it differently.
const divider = (divisor: number) => (dividend: number) => {
  if (divisor === 0) throw new Error('cannot divide by zero')
  return dividend / divisor
}
```

Let's add a division, first with a non-zero divisor:

```typescript
const n = Either.from(4)
  .map(duplicate)
  .map(square)
  .map(flipSign)
  .map(divider(2))

console.log(n) // -> Either { type: 'Right', value: -32 }
```

Now let's try to divide by zero and see what our Either looks like;

```typescript
const n = Either.from(4)
  .map(duplicate)
  .map(square)
  .map(flipSign)
  .map(divider(0))

console.log(n)
```

And the output:

```sh
Either {
  type: 'Left',
  value: Error: cannot divide by zero
    ...stack trace
}
```

That's good, we have a `type` _Left_ and the `value` now holds the error.

One of the nice things about `Either`, is that if we had run into an error earlier, subsequent calls to `map` would not cause any issues. They'd just do nothing more than returning the `Either` with its _Left_ value. In the example above, the last `map` call is the offending one, let's move it up and check that the results are the same.


```typescript
const n = Either.from(4)
  .map(divider(0))
  .map(duplicate)
  .map(square)
  .map(flipSign)

console.log(n) // -> Same output as before.
```

That's all fine, but we need a way to get the underlying value of an `Either`. We will introduce an `unwrap` method to do so. Let's add it to our `Either` class:

```typescript
// either.ts

export default class Either<L, R, Type extends "Left" | "Right" = "Right"> {
  ...

  unwrap(): L | R {
    if (this.isLeft()) {
      return this.value as L
    }
    return this.value as R
  }
}
```

In this case, unwrap returns the value, whether that's a _Left_ or a _Right_. There are many other ways to implement it, for example, you could choose to throw if it's a _Left_. But since I'm not a fan of throwing, I'd rather return whatever is there. The caller knows what a _Left_ is, presumably an Error. I could check if the return value of `unwrap` is an `instanceof Error` and handle it that way. We'll build a more realistic example later and introduce some more methods, for now, the important part is that we can call `unwrap` to get the underlying value.

Let's see it in action:

```typescript
// index.ts

const n = Either.from(4)
  .map(duplicate)
  .map(square)
  .map(flipSign)
  .map(divider(2))

// `result` is of type `number | Error`
const result = n.unwrap()

// This is one way we could handle the `Either` being an error.
if (result instanceof Error) {
  console.log(`[ERROR] ${result}`)
} else {
  // Thanks to the conditional check, result is of type `number` here.
  console.log(`[OK] ${result}`)
}
```

Try the code above dividing by zero too.

`unwrap` is nice. Although I'm not totally sold on the idea of having to do the conditional check on the return value to know if we have an error, it's a useful method to have. But what if I don't care too much about handling the error and want a default value instead? We could of course use `unwrap` and assign the default value after checking if we have an error, but I think it would be more convenient to expose a method to handle that. Let's call it `or`:

```typescript
// either.ts

export default class Either<L, R, Type extends "Left" | "Right" = "Right"> {
  ...

  // It takes its own generic type A in case it is not
  // the same type as R.
  or<A>(alt: A): R | A {
    if (this.isLeft()) {
      return alt
    }
    return this.value as R
  }
}
```

Imagine that in our example, we don't care about errors when doing operations on a number, and if any of them fail, we just want to return `NaN`. Then we can use our new `or` method:

```typescript
const n = Either.from(4)
  .map(duplicate)
  .map(square)
  .map(flipSign)
  .map(divider(0))

// If there's an error (as in this case because we divide by zero),
// `result` is going to be `NaN`, otherwise is going to be
// whatever number our operations return.
const result = n.or(NaN)
console.log(result)
```

`unwrap` and `or` provide some useful ways of getting the value out of `Either` once we are done with our operations. But there's one more method I'd like to add, an `either` method, to which we will provide two functions and the decision on which one to run will be done based on the type of `Either` (meaning _Left_ or _Right_). It could look something like this:

```typescript
// either.ts

export default class Either<L, R, Type extends "Left" | "Right" = "Right"> {
  ...

  either(leftHandler: (value: L) => any, rightHandler: (value: R) => any) {
    if (this.isLeft()) {
      return leftHandler(this.value as L)
    } else {
      return rightHandler(this.value as R)
    }
  }
}
```

And we would use it like this:

```typescript
const errHandler = (err: Error) => console.log(`[ERROR] ${err}`)
const okHandler = (num: number) => console.log(`[OK] ${num}`)

const n = Either.from(4)
  .map(duplicate)
  .map(square)
  .map(flipSign)
  .map(divider(2))

n.either(errHandler, okHandler)
```

The code above outputs `[OK] -32`, and if you change it to divide by zero, you'll get `[ERROR] Error: cannot divide by zero`.

With this, we have the basics of our `Either` monad (I think this is a monad, but who knows...), and we can move to a more realistic example.

## Example

We have a csv file with notifications, social media stuff like "someone liked your post". We will read the file, parse it, and present the number of unread notifications to the user.

Firstly, let's create the csv file, I'm going to call it _notificatins.csv_:

```
2022-12-01T05:33:34,123,liked your post,unread,https://site.com/notifications/3801
2022-12-15T13:47:12,456,replied to your comment,read,https://site.com/notifications/1168
2022-12-28T03:50:08,789,followed you,unread,https://site.com/notifications/88
```

Let's remove the contents of `index.ts` and start fresh:

```typescript
// index.ts

import fs from "node:fs/promises"
import Either from "./either"

// Define a Message type that represents the parsed notification
type Message = {
  date: Date
  from: {
    id: number
    username?: string
  }
  body: string
  read: boolean
  link: URL
}

async function main() {
  const rawData = await fs.readFile("notifications.csv", "utf-8")
  // do something with rawData
}
main()
```

We will need a bunch of functions to operate on the data. These functions are not important for our purposes, so we will keep it simple and don't do all the validation we should for production code.

```typescript
// index.ts

...

function splitRows(data: string): string[] {
  return data.split("\n")
}

function removeEmptyLines(rawMessages: string[]): string[] {
  return rawMessages.filter(i => !!i)
}

function parse(rawMessages: string[]): Message[] {
  return rawMessages.map(m => {
    const fields = m.split(',')
    return {
      date: new Date(fields[0]),
      from: { id: Number(fields[1]) },
      body: fields[2],
      read: fields[3] === "read" ? true : false,
      link: new URL(fields[4]),
    }
  })
}

function unread(messages: Message[]): Message[] {
  return messages.filter(m => !m.read)
}

function count(messages: Message[]): number {
  return messages.length
}
```

With those in place, we can create our `Either` and compute the count of unread messages:

```typescript
// index.ts

...

function errHandler() {
  console.log("Sorry, could not parse notifications")
}

function okHandler(value: number) {
  console.log(`You have ${value} unread messages`)
}

function parseData(rawData: string) {
  const unreadMessagesCount = Either.from(rawData)
    .map(splitRows)
    .map(removeEmptyLines)
    .map(parse)
    .map(unread)
    .map(count)

  unreadMessagesCount.either(errHandler, okHandler)
}

async function main() {
  const rawData = await fs.readFile("notifications.csv", "utf-8")
  parseData(main)
}
main()
```

The code above outputs `You have 2 unread messages`. Try messing up one of the rows in the csv file (like changing the URL to an invalid one) and you will see the error instead. Very nice, but what happens if we need to run async code?

## Async

Some of the operations we wish to perform may be async, so let's think how we can deal with them.

You may have noticed that our `Message` type has a `from.username` optional field that we haven't used so far. Here's the `Message` type as a reminder:

```typescript
type Message = {
  date: Date
  from: {
    id: number
    username?: string
  }
  body: string
  read: boolean
  link: URL
}
```

The notifications in the csv have a user ID, so now we want to fetch the username for the given user ID. We will mock this as an async function as it would be an API call or a database read in real life. Let's create an `addUserData` function (as with the rest of the helper functions, let's not worry about the implementation, the only important thing is that it returns a Promise):

```typescript
// index.ts

const sleep = (ms = 500) => new Promise(r => setTimeout(r, ms))

const usersDB: Record<number, string> = {
  123: 'Maria',
  456: 'Bruce',
  789: 'Isabella',
}

async function addUserData(messages: Message[]): Promise<Message[]> {
  for (const m of messages) {
    await sleep()
    m.from.username = usersDB[m.from.id]
  } 
  return messages
}
```

We no longer want to compute the count of unread messages, we want to show the notifications with the username of the sender. Let's give it a try with the tools we have so far:

```typescript
// index.ts

// Either is now wrapping a `Promise<Message[]>`.
async function okHandlerAsync(promised: Promise<Message[]>) {
  const messages = await promised
  messages.forEach(m => console.log(m))
}

function parseDataAsync(rawData: string) {
  const unreadMessages = Either.from(rawData)
    .map(splitRows)
    .map(removeEmptyLines)
    .map(parse)
    .map(unread)
    .map(addUserData) // addUserData is async

  unreadMessages.either(errHandler, okHandlerAsync)
}

async function main() {
  const rawData = await fs.readFile("notifications.csv", "utf-8")
  // parseData(rawData)
  parseDataAsync(rawData)
}
```

This works, but we have to be aware that `Either` is now wrapping a `Promise` as its _Right_ value, this means that our _okHandler_ needs to await the Promise to access the value, we created an `okHandlerAsync` for that purpose.

It also means that we loose the ability to keep chaining `map` calls, unless the operations expect a Promise as an argument. Which makes it kind of inconvenient to work with. We can improve it by introducing a `mapAsync` method on `Either`.

## Map Async

Let's add another method to `Either`:

```typescript
// either.ts

export default class Either<L, R, Type extends "Left" | "Right" = "Right"> {
  ...

  async mapAsync<T>(fn: (x: R) => Promise<T>): Promise<Either<L, T>> {
    if (this.isLeft()) {
      return Either.left(this.value as L)
    }
    try {
      const newVal = await fn(this.value as R)
      return Either.right(newVal)
    } catch (err) {
      return Either.left(err as L)
    }
  }
}
```

Instead of wrapping the _Right_ value in a Promise, `mapAsync` wraps the entire `Either` instance. This is better because we can `await` it and still get an `Either` back, which means we can keep `map`ping operations. We cannot chain them, but we can introduce intermediary variables.

Let's first redo the example above where we wanted to print the messages with username:

```typescript
// index.ts

...

async function parseDataUsingMapAsync(rawData: string) {
  // Since the last operation is a `mapAsync`, we can await the
  // whole chain to get back the `Either`.
  const unreadMessages = await Either.from(rawData)
    .map(splitRows)
    .map(removeEmptyLines)
    .map(parse)
    .map(unread)
    .mapAsync(addUserData)

  // We're inlining the okHandler
  unreadMessages.either(errHandler, (messages: Message[]) => {
    messages.forEach(m => console.log(m))
  })
}

async function main() {
  const rawData = await fs.readFile("notifications.csv", "utf-8")
  // parseData(rawData)
  // parseDataAsync(rawData)
  parseDataUsingMapAsync(rawData)
}
main()
```

This is nicer, and we can even continue our operations after a `mapAsync`, since we get back an `Either` if we await it. Let's go back to showing the number of unread notifications:

```typescript
async function parseDataUsingMapAsync(rawData: string) {
  const unreadMessages = await Either.from(rawData)
    .map(splitRows)
    .map(removeEmptyLines)
    .map(parse)
    .map(unread)
    .mapAsync(addUserData)

  // We can call `.map` again.
  // The only difference is that we cannot chain it right
  // after `mapAsync` because it returns a Promise, not an Either.
  const unreadMessagesCount = unreadMessages.map(count)

  // This okHandler is the first one we wrote, the one that expects
  // a number for argument.
  unreadMessagesCount.either(errHandler, okHandler)
}
```

The code above prints `You have 2 unread messages` as expected.

Here are the entire contents of `either.ts`:

```typescript
export default class Either<L, R, Type extends "Left" | "Right" = "Right"> {
  private readonly type: "Left" | "Right"
  private readonly value: Type extends "Left" ? L : R

  private constructor(value: Type extends "Left" ? L : R, type: "Left" | "Right" = "Right") {
    this.type = type
    this.value = value;
  }

  static right<L = never, V = never>(value: V): Either<L, V> {
    return new Either<L, V, "Left" | "Right">(value, "Right");
  };

  static left<V = never, R = never>(value: V): Either<V, R> {
    return new Either<V, R, "Left" | "Right">(value, "Left");
  };

  static from<E = Error, V = any>(value: V): Either<E, V> {
    return new Either<E, V, "Right">(value, "Right")
  }

  private isLeft(): boolean {
    return this.type === "Left"
  }

  map<T>(fn: (x: R) => T): Either<L, T> {
    if (this.isLeft()) {
      return Either.left(this.value as L)
    }
    try {
      return Either.right(fn(this.value as R))
    } catch (err) {
      return Either.left(err as L)
    }
  }

  async mapAsync<T>(fn: (x: R) => Promise<T>): Promise<Either<L, T>> {
    if (this.isLeft()) {
      return Either.left(this.value as L)
    }
    try {
      const newVal = await fn(this.value as R)
      return Either.right(newVal)
    } catch (err) {
      return Either.left(err as L)
    }
  }

  unwrap(): L | R {
    if (this.isLeft()) {
      return this.value as L
    }
    return this.value as R
  }

  or<A>(alt: A): R | A {
    if (this.isLeft()) {
      return alt
    }
    return this.value as R
  }

  either(leftHandler: (value: L) => any, rightHandler: (value: R) => any) {
    if (this.isLeft()) {
      return leftHandler(this.value as L)
    } else {
      return rightHandler(this.value as R)
    }
  }
}
```

And that's it, our `Either` monad to handle fallible operations in an elegant, and hopefully readable and clear way.
