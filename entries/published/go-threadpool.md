---
title: go-threadpool
published: 2022-12-14
revision: 2022-12-14
tags: go
excerpt: In Go, it's generally fine to run hundreds of thousands (even millions) of goroutines. However, you may need to limit them. One of the ways to do so is by implementing a pool of workers or threadpool.
---

If you have some background with _Node_, as I do, you've probably seen the pattern where you spawn one instance of your program per CPU. Goroutines are not like that, you can run hundreds of thousands of them, even millions. That's because goroutines are **green threads**, not **os threads**. They are handled by Go and not by the OS directly.

Go handles the management of os threads. If you need to control the number of operating system threads allocated, you can use `GOMAXPROCS`. By default, Go programs run with `GOMAXPROCS` set to the number of cores available. So, most of the times, you needn't worry about it.

Today, we will create a threadpool to limit the amount of goroutines that our program will use. We will have a limited pool of goroutines doing some work, there will never be more than _n_ goroutines running concurrently.

To illustrate with a more or less real example and not just goroutines sleeping for some arbitrary time, we will calculate the area of 100k polygons, which we will read from a text file (_polygons.txt_).

This is what a line of _polygons.txt_ looks like:

```
(153,24),(127,29),(95,34),(79,38),(65,42)
```

You may get the full file [here](/blog/assets/polygons.txt).

_If you are not familiar with polygons or how to calculate their area, don't worry, it's not what we are here for, you can just copy paste that part of the code without worrying too much about what it actually does._

Let's create a new project:

```sh
mkdir threadpool
cd threadpool
go mod init threadpool
touch threadpool.go
mv ~/Downloads/polygons.txt . // Assuming you downloaded polygons.txt to ~/Downloads
```

So, first things first, let's write a program to read the file and calculate the polygons' areas, without any concurrency. We will measure it (just some basic measure, not a proper benchmark) and then improve it by introducing a threadpool.

Since we may end up with a couple different examples, let's follow the convention of using a `cmd` folder.

```sh
mkdir -p cmd/single
touch cmd/single/main.go
```

<br />

```go
// cmd/single/main.go
package main

import (
  "bufio"
  "fmt"
  "log"
  "math"
  "os"
  "regexp"
  "strconv"
  "time"
)

// We'll have a struct to represent each point of a polygon.
// Again, don't worry too much about the geometry of things if you're unfamiliar.
type Point2D struct {
  x int
  y int
}

func main() {
  // Record the start time.
  start := time.Now()

  // Open the file for reads.
  file, err := os.Open("polygons.txt")
  if err != nil {
    log.Fatal(err)
  }
  defer file.Close()

  // For each line of the file, call `calcArea` (each line represents a polygon).
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    calcArea(scanner.Text())
  }

  fmt.Printf("Done in %s\n", time.Since(start))
}

// You can treat this as a black box if you are not interested in the calculation
// of a polygon's area, it just does that and it's not important for our example.
func calcArea(pointStr string) {
  r := regexp.MustCompile(`\((\d*),(\d*)\)`)
  points := []Point2D{}
  for _, p := range r.FindAllStringSubmatch(pointStr, -1) {
    x, _ := strconv.Atoi(p[1])
    y, _ := strconv.Atoi(p[2])
    points = append(points, Point2D{x, y})
  }

  area := 0.0
  for i := 0; i < len(points); i++ {
    a, b := points[i], points[(i+1)%len(points)]
    area += float64(a.x*b.y) - float64(a.y*b.x)
  }

  fmt.Println(math.Abs(area) / 2.0)
}
```

We could read the entire file into memory instead of using a stream, depending on the size of the file, you may prefer one or the other. Generally speaking, I prefer working with streams to know my programs can handle files of any size.

It would look something like:

```go
fileBytes, err := os.ReadFile(filename)
if err != nil {
  log.Fatal(err)
}

for _, line := range strings.Split(string(fileBytes), "\n") {
  calcArea(line)
}
```

But I'm keeping the code with the `bufio.Scanner`. If you run it, you should see all the areas being printed to your terminal and a message at the end saying how long it took.

```sh
go run cmd/single/main.go
```

For me, it's around 3.7 seconds, but your results may be totally different. Check it out.

Let's now use a threadpool. We will first imagine the kind of API we want it to expose and then build it, it's a nice design technique in my opinion.

I want to be able to:

- Set the number of threads in the pool. Throughout this article I'm using threads in the goroutine sense, meaning green threads.
- Tell the workers in the pool what work they need to do. For us that's going to be "calculate the area of a polygon".
- Queue work. In our example that would be "feed polygons to the workers".
- Wait for all queued tasks to be finished.

So, let's create another _main.go_ using this yet-to-be-built threadpool:

```sh
mkdir cmd/pool
touch cmd/pool/main.go
```

<br />

```go
// cmd/pool/main.go
package main

import (
  "bufio"
  "fmt"
  "log"
  "math"
  "os"
  "regexp"
  "strconv"
  "threadpool" // this doesn't exist yet, we'll create it in a second.
  "time"
)

type Point2D struct {
  x int
  y int
}

func main() {
  start := time.Now()

  file, err := os.Open("polygons.txt")
  if err != nil {
    log.Fatal(err)
  }
  defer file.Close()

  // Here we create a new threadpool with 1k threads
  // (or workers or goroutines, however you prefer to think about them).
  //
  // With the second argument (`calcArea`) we tell the workers the task that
  // they will need to run each time they receive some work.
  // In our case, they will receive a string representing a polygon,
  // and they need to calculate its area.
  //
  // The threadpool is generic over the kind of "work" it will get.
  // In our case, we will be feeding it polygons represented as strings.
  // Go can infer that by looking at the signature of our `calcArea` function,
  // so we don't need to be explcit about it.
  // pool := threadpool.New[string](1_000, calcArea) // unnecessary `[string]` annotation.
  pool := threadpool.New(1_000, calcArea)

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    // We send some work to the pool.
    // If there are "idle workers", they will pick it up immediately.
    // If not, it will wait until some worker is free.
    pool.Queue(scanner.Text())
  }

  fmt.Printf("Done in %s\n", time.Since(start))
}

// Exact same function as before.
func calcArea(pointStr string) {
  r := regexp.MustCompile(`\((\d*),(\d*)\)`)
  points := []Point2D{}
  for _, p := range r.FindAllStringSubmatch(pointStr, -1) {
    x, _ := strconv.Atoi(p[1])
    y, _ := strconv.Atoi(p[2])
    points = append(points, Point2D{x, y})
  }

  area := 0.0
  for i := 0; i < len(points); i++ {
    a, b := points[i], points[(i+1)%len(points)]
    area += float64(a.x*b.y) - float64(a.y*b.x)
  }

  fmt.Println(math.Abs(area) / 2.0)
}
```

For this implementation, I decided that the constructor, the `New` function, would take both the number of threads and the function that the workers need to run. You may prefer to provide the task at a later point. But I figured this was the simplest way of explaining it and creating a brief but somewhat real example.

Let's have a first go at the threadpool implementation:

```go
// threadpool.go
package threadpool

// As stated before, our threadpool is generic over the
// the type of work it receives. That's why `T` is also the
// type of the argument to the Task function.
type ThreadPool[T any] struct {
  // The pool will receive work in this channel.
  inputCh chan T

  // The number of threads/workers/goroutines we want in the pool.
  threads int

  // A wrapper that will pick up work from the `inputCh`
  // and run a user-provided function.
  Task    func(chan T)
}

// New creates a ThreadPool with the given amount of threads.
func New[T any](threads int, task func(T)) *ThreadPool[T] {
  pool := &ThreadPool[T]{
    // This is the key part of our threadpool, the buffered channel.
    // By providing the length of the channel, we limit the amount of goroutines.
    inputCh: make(chan T, threads),
    threads: threads,
  }

  pool.Task = func(inputChannel chan T) {
    // Listen to messages on the input channel.
    for work := range inputChannel {
      task(work)
    }
  }

  // We start as many goroutines as the number of threads requested.
  // Each goroutine will execute the pool.Task, which will listen
  // for work on the input channel and run the given function.
  for i := 0; i < threads; i++ {
    go pool.Task(pool.inputCh)
  }

  return pool
}

// Queue sends work to the pool of workers.
func (tp *ThreadPool[T]) Queue(work T) {
  // Send to the channel the workers are listening to.
  tp.inputCh <- work
}
```

If we `go run cmd/pool/main.go`, we will see a lot of output in our terminal, the area of each of the 100k polygons. It may be hard to tell from this output, but we have a problem. If you run it multiple times, maybe you will notice that sometimes, the _"Done in x"_ message is not the last line that's printed. This is already a hint that something is going wrong.

To visualize the problem better, let's change our input to a much smaller one. Let's create a _mini.txt_ with just 3 polygons in it:

```sh
head -n 3 polygons.txt > mini.txt
```

And we will introduce some arbitrary delay in our `calcArea` function.

```go
// cmd/pool/main.go

...

func main() {
  ...
  file, err := os.Open("mini.txt")
  ...
}

func calcArea(pointStr string) {
  time.Sleep(200 * time.Millisecond)
  ...
}
```

If we run it now, we will see that the only line that gets printed is the _"Done in x"_ one. But we were expecting 3 polygon areas to be printed before that, what's going on?

The problem is that our program, our main goroutine, is not waiting for the goroutines in the threadpool to finish their jobs.

To wait for them, we will introduce a `sync.WaitGroup`. In a nutshell, a WaitGroup allows us to keep count of pending tasks and wait for all of them to finish.

```go
// threadpool.go
package threadpool

// New import
import "sync"

type ThreadPool[T any] struct {
  inputCh chan T
  threads int
  Task    func(chan T, *sync.WaitGroup)
  wg      sync.WaitGroup // we add a WaitGroup to our ThreadPool struct
}

func New[T any](threads int, task func(T)) *ThreadPool[T] {
  pool := &ThreadPool[T]{
    inputCh: make(chan T, threads),
    threads: threads,
  }

  // We tell the WaitGroup how many tasks it needs to keep track of.
  pool.wg.Add(threads)

  pool.Task = func(inputChannel chan T, wg *sync.WaitGroup) {
    // Once a worker is done, it calls `wg.Done()` to let the WaitGroup know
    // that it has finished.
    defer wg.Done()
    for work := range inputChannel {
      task(work)
    }
  }

  for i := 0; i < threads; i++ {
    // We pass a reference to the WaitGroup
    // so that goroutines can call `Done()` on it.
    go pool.Task(pool.inputCh, &pool.wg)
  }

  return pool
}

func (tp *ThreadPool[T]) Queue(work T) {
  tp.inputCh <- work
}

// We expose a new method that will allow clients to wait for all workers
// to finish by leveraging the WaitGroup.
//
// It's very important that we close the channel here, otherwise, the defered
// code in `pool.Task` will not run (ranging over a channel finishes when the
// the channel is closed), and that's the code that reduces the WaitGroup counter.
// Without it, the WaitGroup would wait forever. We would get a deadlock error.
func (tp *ThreadPool[T]) Wait() {
  // We close the channel, signaling to all workers that no more work will be sent.
  close(tp.inputCh)
  // We wait for all workers to finish their tasks.
  tp.wg.Wait()
}
```

In our `main` function we just need to call `pool.Wait()`, right before printing the final message:

```go
// cmd/pool/main.go
...

func main() {
  ...
  // This will block until all workers in the pool finish.
  pool.Wait()

  // Same message we had before
  fmt.Printf("Done in %s\n", time.Since(start))
}

...
```

If you run the code now, you should see the expected output, the three areas and the "_Done in x_" message. You can now switch back to _polygons.txt_ instead of _mini.txt_, and get rid of the `time.Sleep(200 * time.Millisecond)` line in `calcArea`.

Running the code with a pool of 1k threads, takes around 2 secs for me, compared to the ~3.7 of the initial version, that's a nice improvement, and we had fun building a threadpool, a good day at the office.

Before we wrap up, one last thing that might come in handy is checking the number of goroutines we're using. This is helpful to confirm our understanding and you might use it if you play around with this implementation and are not sure the threadpool is keeping the expected cap on the number of goroutines.

It's as simple as calling `runtime.NumGoroutine()`. You just need to be carefull where you place the call. Let's do it before and after the call to `pool.Wait()`:

```go
...

func main() {
  ...
  fmt.Printf("Goroutines: #%d\n", runtime.NumGoroutine())
  pool.Wait()
  fmt.Printf("Goroutines: #%d\n", runtime.NumGoroutine())
  ...
}

...
```

You should see something like:

```
Goroutines: #1001
Goroutines: #1
Done in 2s
```

We have set the threadpool to have 1k threads, our _main_ func runs in its own goroutine and so we end up with **1001**. After waiting for all the work to be finished, the goroutines go down to just one (for our _main_). Everything checks out, even though we have 100k polygons, we use 1k goroutines, because that's the limit we set when creating the threadpool.

One other thing that may not be obvious, is that we always allocate 1k goroutines. So, if we are processing a file with 20 polygons in it, it's overkill to have such a large threadpool, 980 goroutines will have nothing to do.
