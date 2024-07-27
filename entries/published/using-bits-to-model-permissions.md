---
title: using-bits-to-model-permissions
published: 2023-01-26
revision: 2023-01-26
tags: ts
excerpt: While looking for an excuse to use bitwise operators and do some bit manipulation, I thought it would be nice to see how we could model permissions and roles using bits.
---

The first thing to keep in mind is that we want a somewhat complex permissions system.
And by that I mean that _higher level_ permissions do not necessarily include the _lower level_ ones.
For example, if we have _user_ and _admin_ roles, generally, the _admin_ would have all the _user_
permissions and more.

The reason we want to avoid this, is that for such a simple scenario, it would
be enough to assign a higher number to the _higher level_ permission and then we
just need to check that the number of the actual permission is greater than or
equal to the minimum required permission to perform a given action.

We could of course model such scenario with bits, but it's more fun when the
system is a bit more complex. So, we will build a model in which permissions are
independent of each other, you'll see what I mean in a minute.

In the context of an accounting software, let's imagine there are four possible
actions:

- issue_invoice
- process_collection
- release_payment
- write_entry

Based on the actions above, we will create five permissions:

- `NONE` -> no permissions at all, can't perform any actions.
- `INVOICER` -> permission to **issue_invoice**, can issue invoices to customers.
- `COLLECTOR` -> permission to **process_collection**, can process payments from customers.
- `PAYER` -> permission to **release_payment**, can make payments to vendors.
- `COOK` -> permission to **write_entry**, can (over)write journal entries.

In this case, since we have 4 actions, we will use 4 bits to represent a
permission. On top of permissions, we will have roles. A role is simply a group
of one or more permissions. We will see this in more detail shortly, but as an example,
if a role had permissions to **release_payment** and **process_collection**, it
would look like:

```
  0 1 1 0
  | | | |
  | | | |__ it does not have permission to `issue_invoice`
  | | |
  | | |__ it has permission to `process_collection`
  | |
  | |__ it has permission to `release_payment`
  |
  |__ it does not have permission to `write_entry`
```

As you probably suspected, `0` means no permission and `1` means go ahead.
Which bit of the 4 we assigned to each of the permissions/actions
has no importance, I just placed them in the same order I had listed them
before (hopefully).

Let's have an enum where we can hold the permissions. As a reminder, in JS, we can type binary
numbers prepending them with `0b`.

_By the way, instead of a typescript `Enum`, I will be using a plain object,
just to show an alternative._

```typescript
// permission.ts

export const Permissions = {
  NONE: 0b0000,
  INVOICER: 0b0001,
  COLLECTOR: 0b0010,
  PAYER: 0b0100,
  COOK: 0b1000,
} as const

// We can extract the keys of the Permissions object into their own type.
// we won't be using it here, but could be useful in other cases
// type PermissionKeys = keyof typeof Permissions

// Extract the values of the Permissions object into their own type.
type Values<T> = T[keyof T]
export type Permission = Values<typeof Permissions>
```

We have defined our `Permissions`, we could also have a type for `Role`. A
`Role` is just going to be an aggregation of `Permissions`, which means that
it will just be a number, but still I think that creating a type for it is going
to make thing clearer and better express intent, so let's do it, and let's also
export a function to facilitate the creation of `Role`s.

```typescript
// permission.ts

...

export type Role = number


// The single `|` character is the bitwise "or" operator.
//     it returns 1 if any of the argument is 1,
//     it returns 0 otherwise.
//
// If we "bitwise or" Permissions.COLLECTOR and Permissions.PAYER, we get
//     - Permissions.COLLECTOR  ->  0 0 1 0
//     - Permissions.PAYER      ->  0 1 0 0
//     - Result (bitwise or)    ->  0 1 1 0
export function createRole(...permissions: Permission[]): Role {
  return permissions.reduce((acc: number, p: number) => acc | p, 0)
}
```

The final piece our _permission_ module is missing, is a helper to create
functions that check that a given `Role` has the correct `Permission` to perform
a certain action.

We've established that a `Role` is a group of `Permissions`, so, when it comes
to checking if a `Role` has certain `Permission`, we just need to check the bit
that corresponds to that `Permission`. If the bit is `1`, we know it is
authorized; if it is `0`, it is not.

How to check for the value of a specific bit you ask? We use the "bitwise and" (`&`).

_Bitwise and_ will return `1` only when both arguments are `1`. So we just need
to compare the actual `Role` with the required `Permission`. The required
`Permission` will of course have the bit we're looking for set to `1` (and the
rest set to `0`). If the actual `Role` has that same bit set to `1`, then we know
that `actualRole & requiredPermission` will for sure be greater than zero,
because it will have one of its bits set to `1`.

Whether the resulting number (in decimal) is `1`, `2`, `4` or `8` will depend
on the bit we are checking, but in all cases we know it's going to be `> 0`,
so we will use that.

```
// permission.ts

...

export function satisfy(p: Permission): (r: Role) => boolean {
  return (r: Role) => (r & p) > 0
}
```

Let's create an `index.ts` file and put our _permission_ module to use,
hopefully that will clarify what we've done.

```
// index.ts

import type { Role } from "./permission"
import { createRole, satisfy, Permissions } from "./permission"

// Create roles
const anon = createRole(Permissions.NONE)
const jr = createRole(Permissions.INVOICER)
const sr = createRole(Permissions.PAYER, Permissions.COLLECTOR)
const owner = createRole(
  Permissions.COLLECTOR,
  Permissions.PAYER,
  Permissions.INVOICER,
  Permissions.COOK
)

// Create some functions to verify permissions
const canInvoice = satisfy(Permissions.INVOICER)
const canCollect = satisfy(Permissions.COLLECTOR)
const canPay = satisfy(Permissions.PAYER)
const canCook = satisfy(Permissions.COOK)

// Let's create a function to log all permissions
function checkAllPermissions(role: Role, label: string) {
  console.log(`===== ${label} =====`)
  console.log(`Invoicer? ${canInvoice(role)}`)
  console.log(`Collector? ${canCollect(role)}`)
  console.log(`Payer? ${canPay(role)}`)
  console.log(`Cook? ${canCook(role)}`)
  console.log()
}

checkAllPermissions(anon, "Anonymous")
checkAllPermissions(jr, "Junior employee")
checkAllPermissions(sr, "Senior employee")
checkAllPermissions(owner, "Owner")
```

If we run `index.ts`, we will get something like:

```
===== Anonymous =====
Invoicer? false
Collector? false
Payer? false
Cook? false

===== Junior employee =====
Invoicer? true
Collector? false
Payer? false
Cook? false

===== Senior employee =====
Invoicer? false
Collector? true
Payer? true
Cook? false

===== Owner =====
Invoicer? true
Collector? true
Payer? true
Cook? true
```

.
