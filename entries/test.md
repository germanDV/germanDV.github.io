---
title: first-post
published: 2022-10-18
revision: 2022-10-20
---

# This is the title

This is inline code `fs := http.FileServer(http.Dir("./static"))`.

Paragraph 1. Obcaecati distinctio blanditiis tempora. **Deserunt magnam, assumenda ab corporis natus ipsam odit libero culpa**. Iure, recusandae ex! Eaque totam mollitia voluptatibus quibusdam veritatis, alias quis, fuga, id odit aperiam facilis?

Paragraph 2. Ex earum blanditiis esse error accusantium natus ducimus fuga! Voluptatem consequuntur repudiandae nihil quae, at numquam est architecto neque odit. Earum optio recusandae provident, aut placeat mollitia fugit cum quos. Pariatur, quod! Blanditiis illo velit officia quod molestiae libero nisi, minima quia. _Itaque sapiente suscipit, similique fugit voluptatum consequatur ea corrupti ex omnis illo, modi deserunt praesentium mollitia nemo dignissimos_.

Paragraph 3. Tiene este link a [debian](https://debian.org) Lorem ipsum dolor sit amet consectetur, adipisicing elit. Officiis nulla culpa, ratione quibusdam voluptatum eos saepe facere, corporis illo iusto, voluptas commodi veniam blanditiis quidem molestias voluptatem quam sit repellat!

Some Go code:

```go
func main() {
  fs := http.FileServer(http.Dir("./static"))
  http.Handle("/static/", http.StripPrefix("/static/", fs))

  // serveTemplate is defined somewhere else.
  http.HandleFunc("/", serveTemplate)

  log.Print("Server up on :4000")
  err := http.ListenAndServe(":4000", nil)
  if err != nil {
    log.Fatal(err)
  }
}
```

Some TS code:

```typescript
import util from 'node:util'

export default class Secret<T> {
  private value: T

  constructor(value: T) {
    this.value = value
  }

  public expose(): T {
    return this.value
  }

  // Return a string with access denied message, instead of throwing an error.
  public toString(): string {
    return 'Secret value access denied, call `expose()` on it to read it.'
  }

  public toJSON(): { value: string } {
    return { value: this.toString() }
  }

  public [util.inspect.custom](): string {
    return this.toString()
  }
}
```

Some Rust code:

```rust
#[derive(serde::Deserialize)]
pub struct ApplicationSettings {
    #[serde(deserialize_with = "deserialize_number_from_string")]
    pub port: u16,
    pub host: String,
}
```

And that's all for now.
