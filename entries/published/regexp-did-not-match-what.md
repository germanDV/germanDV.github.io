---
title: regexp-did-not-match-what
published: 2023-06-11
revision: 2023-06-11
tags: ts
excerpt: This behaviour of Regular Expressions in JS can drive you crazy debugging. Beware!
---

To set the stage, let's imagine a simple _class_ with some utilities to deal with product codes.
Our product codes are very simple, they consist of three uppercase letters, a '-' and four digits, something like `ABC-1234`.

Let's start with a single method to check if a given string contains a valid product code,
maybe it's a weird requirement but let's go with it.

```typescript
class ProdUtil {
  private static readonly ProdCodeRegExp = /[A-Z]{3}-[0-9]{4}/

  static hasValidCode(str: string): boolean {
    return ProdUtil.ProdCodeRegExp.test(str)
  }
}
```

You then use it to check a few strings:

```typescript
console.log(ProdUtil.hasValidCode("2 units of ABC-1234")) // true
console.log(ProdUtil.hasValidCode("5 units of XYZ-4242")) // true
console.log(ProdUtil.hasValidCode("Out of ABC-123")) // false
```

So far so good.

We have a new requirement, we need the ability to scan some text (an email, a stock report, an invoice, etc.)
and extract a list of product codes. So we adjust our RegExp to match multiple occurrences (add the `g` flag), and add a new method:

```typescript
class ProdUtil {
  // Adding the `global` flag.
  private static readonly ProdCodeRegExp = /[A-Z]{3}-[0-9]{4}/g

  // No changes to this one.
  static hasValidCode(str: string): boolean {
    return ProdUtil.ProdCodeRegExp.test(str)
  }

  // New method.
  static extractProdCodes(text: string): string[] {
    // Let's not worry about duplicates.
    const matches = text.match(ProdUtil.ProdCodeRegExp)
    return matches || []
  }
}
```

We can use our new method to extract the product codes mentioned in a dummy message like so:

```typescript
const msg = "Customer A ordered 3 items with code ABC-1234 and 2 items with code DEF-5678"
console.log(ProdUtil.extractProdCodes(msg)) // ["ABC-1234", "DEF-5678"]
```

Everything seems fine, until you use the `hasValidCode` method again. Let's re-run the example above:

```typescript
console.log(ProdUtil.hasValidCode("2 units of ABC-1234")) // true
console.log(ProdUtil.hasValidCode("5 units of XYZ-4242")) // false!!!!!
console.log(ProdUtil.hasValidCode("Out of ABC-123")) // false
```

Look at the second console.log, it was `true` (as it should) two minutes ago, what just happened?
We didn't even touch the `hasValidCode` method. But we did change the RegExp, we added the `g` flag.
And that "breaks" it.

We don't know what's going on so we just add some more console.logs and try to figure it out:

```typescript
console.log(ProdUtil.hasValidCode("ABC-1234")) // true
console.log(ProdUtil.hasValidCode("ABC-1234")) // false!!!!!
console.log(ProdUtil.hasValidCode("ABC-1234")) // true
console.log(ProdUtil.hasValidCode("ABC-1234")) // false!!!!!
```

Is it toggling the checks? What?! Is it really because of that `g`? Let's remove it and try again.

```typescript
class ProdUtil {
  private static readonly ProdCodeRegExp = /[A-Z]{3}-[0-9]{4}/
  ...
}

console.log(ProdUtil.hasValidCode("ABC-1234")) // true
console.log(ProdUtil.hasValidCode("ABC-1234")) // true
console.log(ProdUtil.hasValidCode("ABC-1234")) // true
console.log(ProdUtil.hasValidCode("ABC-1234")) // true
```

Confirmed. But we want that `g`, otherwise we break the `extractProdCodes` method.

We take a step back and look for some documentation / help.

What's going on here is that the RegExp with `g` remembers what has previously been matched,
and starts from the last matched index on subsequent executions.
Totally unexpected outcome if you didn't know that, and a painful bug to track down.

In our example, we're using `ProdCodeRegExp` in both methods, but technically, the `hasValidCode` method does
not need to be global, in practice I would have two different regular expressions.
But for the sake of exploring, let's say we do need to use the same RegExp in both methods.

One way of avoiding this issue, would be to use a new instance of the RegExp in each invocation,
that way we don't have to worry about what it remembers.

```typescript
class ProdUtil {
  static hasValidCode(str: string): boolean {
    return /[A-Z]{3}-[0-9]{4}/g.test(str)
  }

  static extractProdCodes(text: string): string[] {
    const matches = text.match(/[A-Z]{3}-[0-9]{4}/g)
    return matches || []
  }
}
```

That works, the risk here is that our future clean-coder-self is going to forget about this thing
with global RegExp, is going to see the same RegExp being used in two places and is going to extract into
a common property back again. Re-introducing the bug. Hopefully the second time around it'll be an easy fix.

Alternatively, what we can do, is reset the last index matched by the RegExp. That way it will behave
like the first time it's invoked, and the code will make it clearer that we're being intentional about something,
even if we forget what that something is (or if a colleague doesn't know).

```typescript
class ProdUtil {
  private static readonly ProdCodeRegExp = /[A-Z]{3}-[0-9]{4}/g

  static hasValidCode(str: string): boolean {
    const res = ProdUtil.ProdCodeRegExp.test(str)
    // Reset last matched index to avoid funny business.
    ProdUtil.ProdCodeRegExp.lastIndex = 0
    return res
  }

  static extractProdCodes(text: string): string[] {
    const matches = text.match(ProdUtil.ProdCodeRegExp)
    return matches || []
  }
}
```

As the quote goes:

> Some people, when confronted with a problem, think "I know, I'll use regular expressions." Now they have two problems.

Be safe out there.
