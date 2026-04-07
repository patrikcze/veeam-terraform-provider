package datasources

import (
	"context"
	"fmt"
)

// Helper function to safely extract string values from API response
func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// Helper function to safely extract bool values from API response
func getBoolValue(data map[string]interface{}, key string) bool {
	if value, ok := data[key]; ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// Helper function to safely extract int64 values from API response
func getInt64Value(data map[string]interface{}, key string) int64 {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case int:
			return int64(v)
		case int64:
			return v
		case float64:
			return int64(v)
		}
	}
	return 0
}

func getDataList(data map[string]interface{}) []map[string]interface{} {
	if raw, ok := data["data"]; ok {
		if list, ok := raw.([]interface{}); ok {
			items := make([]map[string]interface{}, 0, len(list))
			for _, item := range list {
				if entry, ok := item.(map[string]interface{}); ok {
					items = append(items, entry)
				}
			}
			return items
		}
	}
	return []map[string]interface{}{}
}

func fetchList(ctx context.Context, getter func(context.Context, string, interface{}) error, endpoint string) ([]map[string]interface{}, error) {
	var asArray []map[string]interface{}
	if err := getter(ctx, endpoint, &asArray); err == nil {
		return asArray, nil
	}

	var wrapped map[string]interface{}
	if err := getter(ctx, endpoint, &wrapped); err != nil {
		return nil, err
	}

	items := getDataList(wrapped)
	if len(items) == 0 {
		return []map[string]interface{}{}, nil
	}

	return items, nil
}

// firstNestedID returns the "id" of the first element in a nested array field.
// Used for APIs that return a list of objects (e.g. "roles") where we want the
// primary role's ID without changing the schema to a list.
func firstNestedID(data map[string]interface{}, key string) string {
	arr, ok := data[key].([]interface{})
	if !ok || len(arr) == 0 {
		return ""
	}
	if entry, ok := arr[0].(map[string]interface{}); ok {
		return getStringValue(entry, "id")
	}
	return ""
}

func normalizeDataSourceID(prefix string, value string) string {
	if value == "" {
		return prefix
	}
	return fmt.Sprintf("%s_%s", prefix, value)
}
