package js

// do whatever it takes to get an int64
func toInt64(value interface{}) int64 {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return int64(v)
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case *int:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *int8:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *int16:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *int32:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *int64:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *uint:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *uint8:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *uint16:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *uint32:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *uint64:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *float32:
		if v == nil {
			return 0
		}
		return int64(*v)
	case *float64:
		if v == nil {
			return 0
		}
		return int64(*v)
	}
	return 0
}
