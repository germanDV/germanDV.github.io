---
title: basic-auth-and-the-browser-login-form
published: 2022-12-07
revision: 2022-12-07
tags: go
excerpt: Browsers already have a login form that we can leverage for simple authentication requirements. Let's explore basic auth.
---

The _Basic_ HTTP authentication scheme is defined in [RFC 7617](https://datatracker.ietf.org/doc/html/rfc7617). Basically, credentials (username and password) are base64 encoded and transmitted in the _Authorization_ header.

It looks something like this:

```
Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
```

Let's go through an example. Imagine I have a `/admin` route in this site. Since it's only me who would be accessing that route, and I don't have any other private resources, I figure basic auth is good enough, and it will save a lot of work.

First thing that happens is that the browser GETs `/admin`. The server must return a **401** and the header `WWW-Authenticate: Basic`. When the browser gets this response, it will display a login prompt, asking users to enter username and password.

After entering the credentials, the browser will concatenate the username and password, separated by a `:` and encode it using base64. So, for example:

```
username: "Aladdin"
password: "open sesame"
becomes:  QWxhZGRpbjpvcGVuIHNlc2FtZQ==
```

The encoded credentials are set in the `Authorization` header, prepended by the word `Basic`, and another request is sent to the same endpoint (`/admin` in our example).

This time, the server sees that the request has an `Authorization` header, so, instead of sending back a **401**, it will decode it and validate the user credentials (remember that they are in the format `username:password`).

If credentials are not valid, then it's up to you how to handle it, you could send another **401** so that the user gets another chance to enter the credentials, and you could keep count of how many attempts were made to eventually stop it.

If credentials are valid, then you can proceed normally, whether that means rendering some HTML content, sending some JSON response, etc.

Before we continue, it's important to point out something that the RFC mentions:

> This scheme is not considered to be a secure method of user authentication unless used in conjunction with some external secure system such as TLS (Transport Layer Security, [RFC5246]), as the user-id and password are passed over the network as cleartext.

These days there's really no reason not to be using TLS, but keep in mind that it's particularly important in this case.

We have enough theory to start working on an example. The following will not be our final approach, so, if you are looking for copy-paste material, keep scrolling.

Let's create a simple server with `/public` and `/private` routes.

```go
package main

import (
  "fmt"
  "net/http"
  "os"
  "time"
)

func main() {
  port, ok := os.LookupEnv("PORT")
  if !ok {
    port = "4004"
  }

  mux := &http.ServeMux{}
  mux.Handle("/public", handlePublic())
  mux.Handle("/private", handlePrivate())

  server := &http.Server{
    Addr:         fmt.Sprintf(":%s", port),
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
    Handler:      mux,
  }

  fmt.Printf("Server up on :%s\n", port)
  err := server.ListenAndServe()
  if err != nil {
    panic(err)
  }
}

func handlePublic() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("PUBLIC\n"))
  }
}

func handlePrivate() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("PRIVATE\n"))
  }
}
```

Setting up an `http.Server` is out of the scope of this post, but still I didn't want to use the default `ServeMux` because I don't think it's a good idea, and it's also important to configure timeouts, event if it is not directly related to the current topic of basic auth.

As you can see, our `/private` route is not private at all at the moment. Let's create a middleware to handle authentication.

```go
func basicAuth(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Get `Authorization` header
    authHeader := r.Header.Get("Authorization")

    // If header is not set or does not start with "Basic ", reply 401
    if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
      w.Header().Add("WWW-Authenticate", "Basic")
      w.WriteHeader(http.StatusUnauthorized)
      w.Write([]byte("This endpoint requires authentication.\n"))
      return
    }

    // Validate credentials (we'll write this function in a second)
    if !validCredentials(authHeader) {
      // Here, if credentials are not valid, we're choosing to let the user retry.
      // In reality, I would probably just return an error, or keep track of the number of retries.
      w.Header().Add("WWW-Authenticate", "Basic")
      w.WriteHeader(http.StatusUnauthorized)
      w.Write([]byte("Bad credentials, try again.\n"))
      return
    }

    // Credentials are valid, proceed to the next handler
    next.ServeHTTP(w, r)
  })
}
```

Now that we have the `basicAuth` middleware, we need to use it to protect our `/private` route.

```go
// Before
mux.Handle("/private", handlePrivate())

// After
mux.Handle("/private", basicAuth(handlePrivate()))
```

In our `validCredentials` function, we need to split the string to get rid of the `Basic ` prefix, decode it (it's in base64), split username and password, and compare them to the correct username and password (that we need to retrieve from somewhere). Let's do it:

```go
func validCredentials(authHeader string) bool {
  parts := strings.Split(authHeader, " ")
  if len(parts) != 2 {
    return false
  }

  decoded, err := base64.StdEncoding.DecodeString(parts[1])
  if err != nil {
    return false
  }

  // We could use `strings.Split()` here. The reason we're using `Cut` is that we
  // are only interested in the first appearance of `:`.
  // This allows us to support `:` in the password (not in the username though, be carefull).
  candidateUsername, candidatePassword, ok := strings.Cut(string(decoded), ":")
  if !ok {
    return false
  }

  // The actual valid credentials may be stored in a database, and the password hashed,
  // so you may need some extra steps here.
  // We will assume that the valid credentials are passed in as environment variables.
  username, okU := os.LookupEnv("BASIC_AUTH_USERNAME")
  password, okP := os.LookupEnv("BASIC_AUTH_PASSWORD")

  // If for some reason the credentials are not set in the environment, we probably want to
  // fail quickly so we can address it immediately. Here let's just panic.
  if !okU || !okP {
    panic("Missing environment variables BASIC_AUTH_USERNAME and/or BASIC_AUTH_PASSWORD")
  }

  if candidateUsername != username || candidatePassword != password {
    return false
  }

  return true
}
```

That should do it. To try it out:

* Start the server setting a username and passowrd, for example: `BASIC_AUTH_USERNAME=Aladdin BASIC_AUTH_PASSWORD="open sesame" go run .`
* Open a browser and go to _http://localhost:4004/public_ just to make sure the basic setup is working, you should see `PUBLIC` being rendered in the browser.
* Now go to _http://localhost:4004/private_, you should get a login prompt, if you dismiss it, you will see the message we wrote above (_"This endpoint requires authentication."_). If you enter bad credentials you will get the prompt again. And if you enter the correct credentials, those matching the environment variables from the first step, you will get the protected resource, in this case, the string `PRIVATE` should be rendered in the browser.

We could leave it here, things are working, but remember I said this was not going to be the final solution? That's because we did more work than necessary, Go already supports basic auth, so we don't need to do the decoding and parsing ourselves, we can leverage the `http.Request.BasicAuth()` method to do a lot of the work for us.

This is how our new `basicAuth` middleware and `validCredentials` function look like:

```go
func basicAuth(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    username, password, ok := r.BasicAuth()
    if !ok {
      w.Header().Add("WWW-Authenticate", "Basic")
      w.WriteHeader(http.StatusUnauthorized)
      w.Write([]byte("This endpoint requires authentication.\n"))
      return
    }

    // Validate credentials
    if !validCredentials(username, password) {
      // There's some duplication here, we could abastract it, but we'll leave it as is.
      w.Header().Add("WWW-Authenticate", "Basic")
      w.WriteHeader(http.StatusUnauthorized)
      w.Write([]byte("Bad credentials, try again.\n"))
      return
    }

    // Credentials are valid, proceed to the next handler
    next.ServeHTTP(w, r)
  })
}

func validCredentials(candidateUsername, candidatePassword string) bool {
  username, okU := os.LookupEnv("BASIC_AUTH_USERNAME")
  password, okP := os.LookupEnv("BASIC_AUTH_PASSWORD")

  // If for some reason the credentials are not set in the environment, we probably want to
  // fail quickly so we can address it immediately. Here let's just panic.
  if !okU || !okP {
    panic("Missing environment variables BASIC_AUTH_USERNAME and/or BASIC_AUTH_PASSWORD")
  }

  if candidateUsername != username || candidatePassword != password {
    return false
  }

  return true
}
```

Everything should still be working the same way. One final thing I'd like to address is the fact that our `validCredentials` function is not the safest, it's vulnerable to timing attacks. When we compare `candidatePassword != password`, the more correct characters we have, the longer it will take the function to return, which means an attacker could eventually discover the password, one character at a time, by meassuring the response time.

To prevent this, Go provides us with `subtle.ConstantTimeCompare`. One caveat, the comparison will be in constant time as long as both inputs are of the same length, so, as an extra security meassure, we will hash the inputs before comparison. It doesn't really matter how we hash them, the important part here is not the hashing algorithm, but the fact that we produce same-length inputs.

With that said, here's the complete final code:

```go
package main

import (
  "crypto/sha256"
  "crypto/subtle"
  "fmt"
  "net/http"
  "os"
  "time"
)

func main() {
  port, ok := os.LookupEnv("PORT")
  if !ok {
    port = "4004"
  }

  mux := &http.ServeMux{}
  mux.Handle("/public", handlePublic())
  mux.Handle("/private", basicAuth(handlePrivate()))

  server := &http.Server{
    Addr:         fmt.Sprintf(":%s", port),
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
    Handler:      mux,
  }

  fmt.Printf("Server up on :%s\n", port)
  err := server.ListenAndServe()
  if err != nil {
    panic(err)
  }
}

func handlePublic() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("PUBLIC\n"))
  }
}

func handlePrivate() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("PRIVATE\n"))
  }
}

func basicAuth(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    username, password, ok := r.BasicAuth()
    if !ok {
      w.Header().Add("WWW-Authenticate", "Basic")
      w.WriteHeader(http.StatusUnauthorized)
      w.Write([]byte("This endpoint requires authentication.\n"))
      return
    }

    // Validate credentials
    if !validCredentials(username, password) {
      w.Header().Add("WWW-Authenticate", "Basic")
      w.WriteHeader(http.StatusUnauthorized)
      w.Write([]byte("Bad credentials, try again.\n"))
      return
    }

    // Credentials are valid, proceed to the next handler
    next.ServeHTTP(w, r)
  })
}

func validCredentials(candidateUsername, candidatePassword string) bool {
  // If for some reason the credentials are not set in the environment, we probably want to
  // fail quickly so we can address it immediately. Here let's just panic.
  username, okU := os.LookupEnv("BASIC_AUTH_USERNAME")
  password, okP := os.LookupEnv("BASIC_AUTH_PASSWORD")
  if !okU || !okP {
    panic("Missing environment variables BASIC_AUTH_USERNAME and/or BASIC_AUTH_PASSWORD")
  }

  return match(candidateUsername, username) && match(candidatePassword, password)
}

// match provides a safe way of comapring strings by using
// `subtle.ConstantTimeCompare()` to avoid timing attacks.
// Inputs are hashed to make sure we compare same-length arguments,
// to avoid leaking any information as comparing slices of different
// length is not done in constant time (it returns early).
func match(a, b string) bool {
  hashA := sha256.Sum256([]byte(a))
  hashB := sha256.Sum256([]byte(b))
  return subtle.ConstantTimeCompare(hashA[:], hashB[:]) == 1
}
```

Something I haven't done here is being specific about the _charset_, when we send the `WWW-Authenticate` header, we could add _UTF-8_ to it like so:

```
w.Header().Add("WWW-Authenticate", `Basic charset="UTF-8"`)
```

Finally, one other concept you may find useful is `realm`. It didn't really make sense in this example, and I didn't want to overly complicate things. The idea is that you can have different "protection spaces", as the standard defines them. The `realm` is just a string that can be used to identify the different spaces and require different credentials for each, which the browser will cache.

```
WWW-Authenticate: Basic realm="foo", charset="UTF-8"
```
