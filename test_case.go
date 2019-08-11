package cucumber

type TestCase interface {
	Set(key string, value interface{})
	Get(key string) interface{}
}

type testCase map[string]interface{}

func (tc testCase) Set(key string, value interface{}) {
	tc[key] = value
}

func (tc testCase) Get(key string) interface{} {
	return tc[key]
}
