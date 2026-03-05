package models

type Schema struct {
	BufferSize   int
	Fields       SchemaFields
	CustomFields map[string][]string
	NonJSON      NonJSONRules
}

type SchemaFields struct {
	Timestamp []string
	Level     []string
	Service   []string
	Message   []string
}

type NonJSONRules struct {
	AppendToPreviousRaw *bool
	CreateEventIfNoLast *bool
}

func DefaultSchema() Schema {
	appendRaw := true
	createIfNoLast := true
	return Schema{
		BufferSize: 50000,
		Fields: SchemaFields{
			Timestamp: []string{"ts", "time", "timestamp"},
			Level:     []string{"level"},
			Service:   []string{"logger", "service", "filename"},
			Message:   []string{"msg", "message"},
		},
		CustomFields: map[string][]string{},
		NonJSON: NonJSONRules{
			AppendToPreviousRaw: &appendRaw,
			CreateEventIfNoLast: &createIfNoLast,
		},
	}
}

func (s Schema) ShouldAppendNonJSONToLast() bool {
	if s.NonJSON.AppendToPreviousRaw == nil {
		return true
	}
	return *s.NonJSON.AppendToPreviousRaw
}

func (s Schema) ShouldCreateEventIfNoLast() bool {
	if s.NonJSON.CreateEventIfNoLast == nil {
		return true
	}
	return *s.NonJSON.CreateEventIfNoLast
}
