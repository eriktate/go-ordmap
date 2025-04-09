# go-ordmap

A simple, generic implementation of an ordered, concurrency-safe map in Go.

## Example

```go
package main

import (
  "fmt"

  "github.com/eriktate/go-ordmap"
)

func main() {
  lotrBirthplaces := ordmap.New[string, string](10)
  om.Set("Frodo", "The Shire")
  om.Set("Legolas", "Mirkwood")
  om.Set("Aragorn", "Arnor")
  om.Set("Gimli", "The Blue Mountains")

  for name, birthplace := range lotrCharacters.EntryIter() {
    fmt.Printf("%s is from %s\n", name, birthplace)
  }
}
```

This is guaranteed to always print exactly the following:

```
Frodo is from The Shire
Legolas is from Mirkwood
Aragorn is from Arnor
Gimli is from The Blue Mountains
```
