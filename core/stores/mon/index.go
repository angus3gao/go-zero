package mon

type Index struct {
	Collection string
	Unique     bool
	Fields     []string
}

func NewIndex(collection string, unique bool, fields ...string) *Index {
	return &Index{
		Collection: collection,
		Unique:     unique,
		Fields:     fields,
	}
}
