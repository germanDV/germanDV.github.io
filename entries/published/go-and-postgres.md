---
title: go-and-postgres
published: 2023-10-29
revision: 2023-10-29
excerpt: Working with PostgreSQL in Go, using the pgx library.
---

In this article, we will be working with PostgreSQL in Go, using the [pgx](https://github.com/jackc/pgx) library.
Pgx is a PostgreSQL driver, recommended by **lib/pq**, which is now in maintenance mode.
The pattterns recommended here are inspired by
[this talk](https://www.youtube.com/watch?v=sXMSWhcHCf8) by pgx's author.


## Setup

We'll be running Postgres in a Docker container. This will do:

```sh
docker run \
    --name gopgx-postgres \
    -e POSTGRES_PASSWORD=pass123  \
    -e POSTGRES_USER=postgres  \
    -e POSTGRES_DB=example \
    -p 5432:5432 \
    -d postgres:15-alpine
```

We'll call the project gopgx:

```sh
go mod init gopgx
```

To keep the example focused on pgx, We won't worry about project structure, it will all be in the
`main` package. And We'll skip error handling for brevity.

Let's create a db.go file with the connection logic and some initial data for the example:

```go
// db.go

package main

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

// connect establishes a PostgreSQL connection pool.
func connect(dsn string) (*pgxpool.Pool, error) {
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, err
    }

    cfg.MaxConnIdleTime = 15 * time.Minute

    pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = pool.Ping(ctx)
    if err != nil {
        return nil, err
    }

    return pool, nil
}

// seed creates tables and populates them.
func seed(db *pgxpool.Pool) error {
  create := `
    create table if not exists squads(
      id serial primary key,
      name text not null unique
    );

    create table if not exists employees (
      id serial primary key,
      fname text not null,
      lname text not null,
      position text not null,
      squad_id int references squads(id)
    );

    create table if not exists wallets(
      id serial primary key,
      address text not null,
      owner_id int references employees(id) unique
    );
  `

  insert := `
    insert into squads (id, name) values
       (1, 'Microservice Nonsense'),
       (2, 'Yaml Developers');

    insert into employees (id, fname, lname, position, squad_id) values
       (1, 'John', 'Doe', 'FrontEnd Developer', 1),
       (2, 'Jane', 'Doe', 'FullStack Developer', 1),
       (3, 'Joe', 'Smith', 'BackEnd Developer', 1),
       (4, 'Lois', 'Lang', 'BackEnd Developer', 2),
       (5, 'Keira', 'Gordon', 'FrontEnd Developer', 2);

    insert into wallets (id, address, owner_id) values
      (1, '0x1234567890123456789012345678901234567891', 1),
      (2, '0x1234567890123456789012345678901234567892', 2),
      (3, '0x1234567890123456789012345678901234567893', 3),
      (4, '0x1234567890123456789012345678901234567894', 4),
      (5, '0x1234567890123456789012345678901234567895', 5);
  `

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  _, err := db.Exec(ctx, create)
  if err != nil {
    return err
  }

  _, err = db.exec(ctx, insert)
    if err != nil {
        return err
  }

  return nil
}
```

This is a good time to run `go mod tidy` so that pgx is installed. I'm using v5, the latest as of the time of writing.

In our `main.go` file we will connect to the DB and seed it:

```go
// main.go

// This are all the imports we're going to need.
import (
    "flag"
    "fmt"
)

func main() {
    db, _ := connect("postgres://postgres:pass123@localhost:5432/example")

    shouldSeed := flag.Bool("seed", false, "Setup & Seed Database")
    flag.Parse()
    if *shouldSeed {
        _ = seed(db)
        return
    }
}
```

We are hardcoding the connection string, in reality you'd probably use environment variables.

If we now run `go run . -seed`, it should successfully connect to the DB, create the tables and
insert some rows. Check your DB to confirm that everything is working as expected.

## Retrieving Data

As you might guess from the tables we created, we will be working with three domains, let's create
some files and define some structs.

Employees:

```go
// employee.go

package main

// This are all the imports we're going to need.
import (
    "context"
    "time"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Employee struct {
    ID        int    `db:"id"`
    FirstName string `db:"fname"`
    LastName  string `db:"lname"`
    Position  string `db:"position"`
    SquadID   int    `db:"squad_id"`
}
```

Squads:
```go
// squad.go

package main

// This are all the imports we're going to need.
import (
    "context"
    "time"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Squad struct {
    ID      int        `db:"id"`
    Name    string     `db:"name"`
    Members []Employee `db:"members"`
}
```

Wallets:

```go
// wallet.go

package main

// This are all the imports we're going to need.
import (
    "context"
    "time"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Wallet struct {
    ID      int
    Address string
    Owner   *Employee
}
```

Our first task will be to get all rows from the `employees` table.

### The database/sql approach

The "standard" way that you've probably seen many times to scan multiple rows would look something like this:

```go
func GetEmployees(db *pgxpool.Pool) ([]Employee, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    employees := []Employee{}

    rows, err := db.Query(ctx, "select id, fname, lname, position, squad_id from employees")
    if err != nil {
        return employees, err
    }

    for rows.Next() {
        e := Employee{}
        err = rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Position, &e.SquadID)
        if err != nil {
            return employees, err
        }
        employees = append(employees, e)
    }

    if rows.Err() != nil {
        return employees, err
    }

    return employees, nil
}
```

### The pgx approach

We can leverage pgx to make the above less verbose and less error-prone.

```go
// employee.go

// ...

func GetEmployees(db *pgxpool.Pool) ([]Employee, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // It is safe to ignore the error here because pgx.CollectRows will handle it.
    rows, _ := db.Query(ctx, "select id, fname, lname, position, squad_id from employees")

    // Here we can use pgx.RowToStructByName because our Employee struct has "db" tags
    // that allow to map column names to struct fields.
    // If we didn't want to use "db" struct tags, we could have used a different pgx method,
    // we'll have an example of that later.
    return pgx.CollectRows(rows, pgx.RowToStructByName[Employee])
}
```

### Run It

To run the code and check the results, we can hook it up in our `main.go` file.

```go
// main.go

// ...

func main() {
    // ...

    fmt.Println("GetEmployees")
    employees, _ := GetEmployees(db)
    for _, employee := range employees {
        fmt.Printf("\t%v\n", employee)
    }
}
```

And then run `go run .` to have all the employees printed to stdout.

_I won't be showing this step of hooking things up in `main.go` from now on, but I encourage you to do so._

Let's write a quick example of how to get a single row:

```go
// employee.go

// ...

func GetEmployee(db *pgxpool.Pool, id int) (Employee, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    rows, _ := db.Query(ctx, "select id, fname, lname, position, squad_id from employees where id = $1", id)
    return pgx.CollectOneRow(rows, pgx.RowToStructByName[Employee])
}
```

We are again using `pgx.RowToStructByName`, the difference now is that we use `pgx.CollectOneRow` instead of `pgx.CollectRows`.

### Nested Structs

Let's now move to a more interesting example, involving nested structs.
The `Squad` struct has a `Members` field, which is a slice of `Employee`.
Our goal is to get all squads with their corresponding members, and we can do so in a single query.

```
// squad.go

// ...

func GetSquad(db *pgxpool.Pool, id int) (Squad, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    q := `select s.id, s.name,
    (
      select array_agg(
        row(
          e.id,
          e.fname,
          e.lname,
          e.position,
          e.squad_id
        )
      )
      from employees e
      where e.squad_id = s.id
    ) as members
    from squads s
    where s.id = $1;
  `

    rows, _ := db.Query(ctx, q, id)
    return pgx.CollectOneRow(rows, pgx.RowToStructByName[Squad])
}
```

With the nested select, we get a result set that pgx can scan into our struct, including the nested
`[]Employee` field.
Since we are using `pgx.RowToStructByName`, it's important that we name the nested select _members_.
Another important thing to notice is that this uses Postgres' extended protocol, which returns type
information. If this doesn't work with a cloud-hosted Postgres instance, keep it in mind.

Of course, we can also scan into a nested struct that's not a slice, and that's a pointer.
We have such scenario in the `Wallet` struct, which has an `Owner` field that is a pointer to an
`Employee` struct.

```go
// wallet.go

// ...

func GetWallet(db *pgxpool.Pool, employeeID int) (Wallet, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    q := `select w.id, w.address, row(e.id, e.fname, e.lname, e.position, e.squad_id) as owner
    from wallets w
    join employees e on w.owner_id = e.id
    where w.owner_id = $1`

    rows, _ := db.Query(ctx, q, employeeID)
    return pgx.CollectOneRow(rows, pgx.RowToStructByPos[Wallet])
}
```

Just to show different options, we are using a **join** instead of a nested **select** in this one.
Also, you may have noticed we used `pgx.RowToStructByPos` instead of `pgx.RowToStructByName`.
In this case, the mapping from column to struct field is done based on the position of the fields,
that's why we don't need any "db" struct tags in `Wallet`.

## Transactions

Let's now imagine that we need the ability to have two employees switch squads.
The switch should happen atomically, meaning, either both employees are transferred successfully or none of them are; we need a transaction.

### The database/sql approach

A transaction may look something like this:

```go
func Switch(db *pgxpool.Pool, empA int, empB int) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    tx, err := db.Begin(ctx)
    if err != nil {
        return err
    }

    // It is safe to call Rollback in a committed transaction.
    // We can take advantage of this behaviour to defer the rollback instead of having to
    // manually call it at every error occurence.
    defer tx.Rollback(ctx)

    q := "select squad_id from employees where id = $1"

    var squadA int
    err := tx.QueryRow(ctx, q, empA).Scan(&squadA)
    if err != nil {
        return err
    }

    var squadB int
    err = tx.QueryRow(ctx, q, empB).Scan(&squadB)
    if err != nil {
        return err
    }

    q = "update employees set squad_id = $1 where id = $2"

    _, err = tx.Exec(ctx, q, squadB, empA)
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, q, squadA, empB)
    if err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

### The pgx approach

pgx provides a `BeginFunc` method that automatically starts a transaction and commits it if there
are no errors, or rolls it back if there are any.

```go
// employee.go

// ...

func Switch(db *pgxpool.Pool, empA int, empB int) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
        // If this function returns an error, the tx is reverted, otherwise the tx is committed.
        q := "select squad_id from employees where id = $1"

        var squadA int
        err := tx.QueryRow(ctx, q, empA).Scan(&squadA)
        if err != nil {
            return err
        }

        var squadB int
        err = tx.QueryRow(ctx, q, empB).Scan(&squadB)
        if err != nil {
            return err
        }

        q = "update employees set squad_id = $1 where id = $2"

        _, err = tx.Exec(ctx, q, squadB, empA)
        if err != nil {
            return err
        }

        _, err = tx.Exec(ctx, q, squadA, empB)
        if err != nil {
            return err
        }

        return nil
    })
}
```

Thanks to the possibility of `defer tx.Rollback()`, I personally don't see `pgx.BeginFunc` as a huge
improvement or a much better option, but one could argue it is more readable since deferring the
rollback could look strange, and some might like that they don't have to manually call commit.

## End

I have handpicked a few things from pgx that I think are especially useful and not necessarily super clear in the documentation.
So hopefully this is helpful even for developers who already had some experience working with it.
