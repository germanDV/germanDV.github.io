---
title: utility-to-deal-with-secrets-in-TS
published: 2022-11-30
revision: 2022-11-30
tags: ts
excerpt: If your code deals, at one point or another, with secrets in plain text, it might be a good idea to prevent accidental logging of such sensitive information.
---

This is an adaptation of [the post about secrets in Go](./utility-to-deal-with-secrets-in-go.html) that can be used in Node applications, whether using TS or JS.

In summary, the idea is to add some level of protection to plain text secrets that you may have in your code, like API keys or user passwords that you need to hash and compare with the ones in your database.

By wrapping values in a `Secret`, you convey the secrecy of the value to anyone reading the code, avoid accidentally logging it, and have to be explicit about accessing it.

The implementation:

```typescript
import util from 'node:util'

class Secret<T> {
  private value: T

  constructor(value: T) {
    this.value = value
  }

  public expose(): T {
    return this.value
  }

  public toString(): string {
    return 'Secret value access denied, call `expose()` on it to read it.'
  }

  public toJSON(): { value: string } {
    return { value: this.toString() }
  }

  // This is necessary to tell `console.log` to use our `toString()` method
  // and prevent it from logging the Secret instance as any other object,
  // which would reveal our secret value.
  public [util.inspect.custom](): string {
    return this.toString()
  }
}
```

The usage:

```typescript
type User = {
  email: string
  password: Secret<string>
}

const u: User = {
  email: 'user@service.tld',
  password: new Secret('P4ssw0rd!'),
}

// None of these will expose the password.
console.log(u)
console.log(u.password)
console.log(JSON.stringify(u))

// Not allowed by TS.
// console.log(u.password.value)

// Correct way of getting the underlying value back.
console.log(u.password.expose())
```

.
