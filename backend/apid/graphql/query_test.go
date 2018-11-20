package graphql

// TODO: Fixup tests
// func TestQueryTypeEventField(t *testing.T) {
// 	mock := mockEventFetcher{&types.Event{Timestamp: 123456789}, nil}
// 	impl := queryImpl{eventFinder: mock}
//
// 	args := schema.QueryEventFieldResolverArgs{Namespace: "ns"}
// 	params := schema.QueryEventFieldResolverParams{Args: args}
//
// 	res, err := impl.Event(params)
// 	require.NoError(t, err)
// 	assert.NotEmpty(t, res)
// }
//
// func TestQueryTypeNamespaceField(t *testing.T) {
// 	mock := mockNamespaceFinder{&types.Namespace{Name: "us-west-2"}, nil}
// 	impl := queryImpl{nsFinder: mock}
//
// 	params := schema.QueryNamespaceFieldResolverParams{}
// 	params.Args.Name = "us-west-2"
//
// 	res, err := impl.Namespace(params)
// 	require.NoError(t, err)
// 	assert.NotEmpty(t, res)
// }
//
// func TestQueryTypeEntityField(t *testing.T) {
// 	mock := mockEntityFetcher{types.FixtureEntity("abc"), nil}
// 	impl := queryImpl{entityFinder: mock}
//
// 	params := schema.QueryEntityFieldResolverParams{}
// 	params.Args.Namespace = "ns"
// 	params.Args.Name = "abc"
//
// 	res, err := impl.Entity(params)
// 	require.NoError(t, err)
// 	assert.NotEmpty(t, res)
// }
