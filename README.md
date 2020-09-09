elsearm
===========
[![CircleCI](https://circleci.com/gh/soranoba/elsearm.svg?style=svg&circle-token=ad0136e25df22988c8370d90926a8c4f1fb9fc61)](https://circleci.com/gh/soranoba/elsearm)
[![Go Report Card](https://goreportcard.com/badge/github.com/soranoba/elsearm)](https://goreportcard.com/report/github.com/soranoba/elsearm)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/soranoba/elsearm)](https://pkg.go.dev/github.com/soranoba/elsearm)

elsearm is an **El**astic**sear**ch **m**odel library for Go.

## Features

- Only has [elastic/go-elasticsearch](https://github.com/elastic/go-elasticsearch) as a dependency.
- Easy update/remove to Elasticsearch when saving to DB.
- Can set prefix and suffix of index name.

## Usage

### Automatic Document Updates (with [GORM](https://github.com/go-gorm/gorm))

If the ORM library support hooks, you can use **Automatic Document Updates** with libraries other than GORM.<br>
This is an example of using GORM.

```go
package models

import (
	"github.com/soranoba/elsearm"
	"gorm.io/gorm"
)

const (
	elsearmIndexerKey = "elsearm:indexer"
)

func UseElsearmIndexer(db *gorm.DB, indexer *elsearm.Indexer) *gorm.DB {
	return db.InstanceSet(elsearmIndexerKey, indexer)
}

func getElsearmIndexer(db *gorm.DB) (*elsearm.Indexer, bool) {
	val, ok := db.InstanceGet(elsearmIndexerKey)
	if !ok {
		return nil, false
	}
	indexer, ok := val.(*elsearm.Indexer)
	return indexer, ok
}

type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) AfterDelete(db *gorm.DB) error {
	if indexer, ok := getElsearmIndexer(db); ok {
		indexer.Delete(u)
	}
	return nil
}

func (u *User) AfterSave(db *gorm.DB) error {
	if indexer, ok := getElsearmIndexer(db); ok {
		indexer.Update(u)
	}
	return nil
}
```

```go
func main() {
	db := (func() *gorm.DB {
		return /* do anything */
	})()
	indexer := elsearm.NewIndexer(ct.es)
	db = models.UseElsearmIndexer(db, indexer)

	/* do anyting */
	db.Create(&models.User{})
}
```

### Using another index name with tests.

You can set prefix and/or suffix in global config.<br>
It makes it easy to avoid overwrite to the same index when it execute testings.

```go
elsearm.SetGlobalConfig(elsearm.GlobalConfig{
	IndexNameSuffix: "_test",
})
```
