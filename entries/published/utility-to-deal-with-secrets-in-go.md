---
title: utility-to-deal-with-secrets-in-go
published: 2022-11-29
revision: 2022-11-29
excerpt: If your code deals, at one point or another, with secrets in plain text, it might be a good idea to prevent accidental logging of such sensitive information. 
---



Maybe you have a user password in a string during the login/signup process, or an API key that you read from the environment, and you inadvertently log it somewhere. 
Even if your logs are private, you may not want to disclose sensitive information.
There might even be regulations that prevent you from keeping certain information in your system.
Or maybe you are printing some struct for debugging purposes and would like to share the output with someone else to ask for help, but you may have secret information there, that would be more appropriate to leave out.

One simple way around this is to wrap secrets in a `Secret` _struct_ that prints an error message instead of the wrapped value; requiring you to explicitly _expose_ the underlying value whenever you wish to access it.

Let's start by defining the `Secret` struct, it just takes the value it needs to wrap. The important part is that `value` is not exported (lowercased). If you create a _secret_ package inside your application, this will ensure that no one outside that package will be able to access the value with `.value`.

```go
type Secret[T any] struct {
    value T
}
```

Nothing special so far. We need to make sure that trying to log `Secret` does not expose `value`.

The first thought could be for `Secret` to implement the [Stringer](https://pkg.go.dev/fmt#Stringer) interface:

```go
type Stringer interface {
    String() string
}
```

For example like so:

```go
func (s Secret[T]) String() string {
    return "Secret value access denied."
}
```

Now, when we try to log a `Secret`, we get our warning instead:

```go
pass := Secret[string]{"secret_api_key"}
fmt.Println(pass)
fmt.Printf("Secret is: %s\n", pass)
fmt.Printf("Secret is: %q\n", pass)
str := fmt.Sprint(pass)
fmt.Println(str)
```

In all cases above, we will get _"Secret value access denied."_ instead of `secret_api_key`. We are making progress, but we are not there yet. The `v` verb with the `#` flag will actually print our secret value, that's because in this case, the `GoString` method is used, not the `String` one.

```go
pass := Secret[string]{"secret_api_key"}
fmt.Printf("%#v\n", pass)
```

The code above prints something like `main.Secret[string]{value:"secret_api_key"}`, revealing our secret.

To cover this scenario, we could also implement the [GoStringer](https://pkg.go.dev/fmt#GoStringer) interface:

```go
type GoStringer interface {
    GoString() string
}
```

We can simply do the same thing we did for `Stringer`:

```go
func (s Secret[T]) GoString() string {
    return "Secret value access denied."
}
```

Now, the snippet above, using `%#v` will also print the _"Secret value access denied."_ message. Nice. However, there's one edge case.

If we do something silly like trying to cast the value (which is a _string_ in this example) to another type, it will expose the secret:

```go
pass := Secret[string]{"secret_api_key"}
fmt.Printf("%d\n", pass) // -> {%!d(string=secret_api_key)}
```

Your text editor should warn you about this "wrong type" issue, if you run `go vet`, you should see something like _fmt.Printf format %d has arg pass of wrong type command-line-arguments.Secret[string]_, but the code will still compile and run.

At this point, you may be happy leaving things as they are, after all, you should be vetting and linting your code, catching and fixing these warnings. However, there's an easy fix, so let's implement it.

Instead of implementing the `Stringer` and `GoStringer` interfaces, we could go one level deeper and implement the [Formatter](https://pkg.go.dev/fmt#Formatter) interface:

```go
type Formatter interface {
    Format(f State, verb rune)
}
```

We can, for example, implement it like so:

```go
func (s Secret[T]) Format(f fmt.State, verb rune) {
    fmt.Fprint(f, "Secret value access denied.")
}
```

This takes care of all our needs, so we can remove the `String` and `GoString` methods. This is the entire code so far:

```go
package main

import (
    "fmt"
)

type Secret[T any] struct {
    value T
}

func (s Secret[T]) Format(f fmt.State, verb rune) {
    fmt.Fprint(f, "Secret value access denied.")
}

func main() {
    pass := Secret[string]{"secret_api_key"}
    fmt.Println(pass)                     // -> Secret value access denied.
    fmt.Printf("%#v\n", pass)             // -> Secret value access denied.
    fmt.Println(fmt.Sprintf("%s", pass))  // -> Secret value access denied.
    fmt.Printf("%d\n", pass)              // -> Secret value access denied.
}
```

Now, our secret is protected, but it's not very useful if can't get it back, let's add an `Expose` method that we will call explicitly when we actually need the secret value:

```go
// Expose returns the wrapped secret value.
func (s Secret[T]) Expose() T {
    return s.value
}
```

Now, we can get the secret value by calling `Expose()` on it:

```go
fmt.Println(pass.Expose()) // -> secret_api_key
```

Finally, we will create a separate package for our `Secret`, and make some smaller adjustments like providing a constructor function and adjusting the warning message:

```go
package secret

import "fmt"

type Secret[T any] struct {
    value T
}

// New wraps the provided value in a `Secret` and returns it.
func New[T any](v T) Secret[T] {
    return Secret[T]{value: v}
}

func (s Secret[T]) Format(f fmt.State, verb rune) {
    fmt.Fprint(f, "Secret value access denied, call `expose()` to read it.")
}

// Expose returns the wrapped secret value.
func (s Secret[T]) Expose() T {
    return s.value
}
```

And our `main.go`:

```go
package main

import (
    "example/secret"
    "fmt"
)

func main() {
    pass := secret.New("secret_api_key")

    // All these will return `Secret value access denied, call `expose()` to read it.`
    fmt.Printf("%d\n", pass)
    fmt.Println(pass)
    fmt.Printf("%#v\n", pass)
    fmt.Println(fmt.Sprintf("%s", pass))

    // Cannot access unexported field.
    // fmt.Println(pass.value)

    // This is the correct way of accessing the wrapped value.
    fmt.Println(pass.Expose())
}
```

In addition to preventing some accidental exposure of secrets, I like this approach because anyone reading the code will immediately understand that we are dealing with sensitive information.

Hope you find it useful.
