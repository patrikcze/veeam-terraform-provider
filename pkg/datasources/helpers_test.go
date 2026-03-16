package datasources

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchList_FromArray(t *testing.T) {
	getter := func(_ context.Context, _ string, result interface{}) error {
		array, ok := result.(*[]map[string]interface{})
		if !ok {
			return errors.New("expected array")
		}
		*array = []map[string]interface{}{{"id": "1"}, {"id": "2"}}
		return nil
	}

	items, err := fetchList(context.Background(), getter, "/api/v1/test")
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "1", getStringValue(items[0], "id"))
}

func TestFetchList_FromWrappedData(t *testing.T) {
	getter := func(_ context.Context, _ string, result interface{}) error {
		switch target := result.(type) {
		case *[]map[string]interface{}:
			return errors.New("not an array response")
		case *map[string]interface{}:
			*target = map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": "a"},
					map[string]interface{}{"id": "b"},
				},
			}
			return nil
		default:
			return errors.New("unsupported target")
		}
	}

	items, err := fetchList(context.Background(), getter, "/api/v1/test")
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "a", getStringValue(items[0], "id"))
}

func TestNormalizeDataSourceID(t *testing.T) {
	assert.Equal(t, "prefix", normalizeDataSourceID("prefix", ""))
	assert.Equal(t, "prefix_value", normalizeDataSourceID("prefix", "value"))
}
