package helper

func Bool(i bool) *bool { return &i }

func String(i string) *string { return &i }

func Int(i int) *int { return &i }

func Uint(i uint) *uint { return &i }

func Int64(i int64) *int64 { return &i }

func Uint64(i uint64) *uint64 { return &i }

func IntInt64(i int) *int64 {
	i64 := int64(i)
	return &i64
}

func IntUint64(i int) *uint64 {
	u := uint64(i)
	return &u
}

func Strings(strs []string) []*string {
	if len(strs) == 0 {
		return nil
	}

	sp := make([]*string, 0, len(strs))
	for _, s := range strs {
		sp = append(sp, String(s))
	}

	return sp
}

func PString(pointer *string) string {
	if pointer == nil {
		return ""
	}
	return *pointer
}

// Takes the result of flatmap.Expand for an array of strings
// and returns a []string
func InterfacesStrings(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, v.(string))
	}
	return vs
}

// Flatten to an array of raw strings and returns a []interface{}
func StringsInterfaces(list []*string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, *v)
	}
	return vs
}

func Uint64sInterfaces(list []*uint64) []interface{} {
	vi := make([]interface{}, 0, len(list))
	for _, v := range list {
		vi = append(vi, int(*v))
	}
	return vi
}