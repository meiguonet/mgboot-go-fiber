package cachex

type noopCache struct {
}

func (c *noopCache) Get(_ string, _ ...interface{}) interface{} {
	return nil
}

func (c *noopCache) Set(_ string, _ interface{}, _ ...interface{}) bool {
	return false
}

func (c *noopCache) Delete(_ string) bool {
	return false
}

func (c *noopCache) Clear() bool {
	return false
}

func (c *noopCache) GetMultiple(_ []string, _ ...interface{}) []interface{} {
	return make([]interface{}, 0)
}

func (c *noopCache) SetMultiple(_ []map[string]interface{}, _ ...interface{}) bool {
	return false
}

func (c *noopCache) DeleteMultiple(_ []string) bool {
	return false
}

func (c *noopCache) Has(_ string) bool {
	return false
}
