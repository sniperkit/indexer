package elastic_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/dotnews/indexer/config"
	"github.com/dotnews/indexer/elastic"
	"github.com/stretchr/testify/assert"
)

var c = config.New("../")
var el = elastic.New(c)

func TestMain(m *testing.M) {
	exists, err := el.Client.IndexExists(c.Article.Index).Do(el.Context)
	if err != nil {
		panic(err)
	}

	if exists {
		res, err := el.Client.DeleteIndex(c.Article.Index).Do(el.Context)
		if err != nil {
			panic(err)
		}
		if !res.Acknowledged {
			panic(fmt.Sprintf("Index deletion was not acknowledged: %+v", res))
		}
	}

	os.Exit(m.Run())
}

func TestCreate(t *testing.T) {
	err := el.Create(c.Article.Index, "{}")
	assert.Nil(t, err)

	exists, err := el.Client.IndexExists(c.Article.Index).Do(el.Context)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestEnsure(t *testing.T) {
	res, err := el.Client.DeleteIndex(c.Article.Index).Do(el.Context)
	assert.Nil(t, err)
	assert.True(t, res.Acknowledged)

	for i := 1; i <= 2; i++ {
		err = el.Ensure(c.Article.Index, "{}")
		assert.Nil(t, err)

		exists, err := el.Client.IndexExists(c.Article.Index).Do(el.Context)
		assert.Nil(t, err)
		assert.True(t, exists)
	}
}

func TestDelete(t *testing.T) {
	err := el.Delete(c.Article.Index)
	assert.Nil(t, err)

	exists, err := el.Client.IndexExists(c.Article.Index).Do(el.Context)
	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestIndex(t *testing.T) {
	res, err := el.Client.CreateIndex(c.Article.Index).Do(el.Context)
	assert.Nil(t, err)
	assert.True(t, res.Acknowledged)

	sr, err := el.Client.Search().
		Index(c.Article.Index).
		From(0).
		Size(10).
		Do(el.Context)

	assert.Nil(t, err)
	assert.Equal(t, int64(0), sr.TotalHits())

	err = el.Index(c.Article.Index, c.Article.Type, "id", "{}")
	assert.Nil(t, err)

	_, err = el.Client.Flush(c.Article.Index).Do(el.Context)
	assert.Nil(t, err)

	sr, err = el.Client.Search().
		Index(c.Article.Index).
		From(0).
		Size(10).
		Do(el.Context)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), sr.TotalHits())
}

func TestGet(t *testing.T) {
	err := el.Index(c.Article.Index, c.Article.Type, "id", "{\"foo\": \"bar\"}")
	assert.Nil(t, err)

	b, err := el.Get(c.Article.Index, c.Article.Type, "id")
	var source map[string]interface{}
	err = json.Unmarshal(b, &source)

	assert.Nil(t, err)
	assert.Equal(t, "bar", source["foo"])
}

func TestGetMapping(t *testing.T) {
	s, err := el.GetMapping(c.Article.Mapping)
	assert.Nil(t, err)

	var m map[string]interface{}
	err = json.Unmarshal([]byte(s), &m)
	assert.Nil(t, err)
	assert.True(t, len(s) > 0)
}
