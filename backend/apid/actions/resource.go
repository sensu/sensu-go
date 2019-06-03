package actions

// ResourceController is a generic controller that deals with Resource objects
// type ResourceController struct {
// 	store store.ResourceStore
// 	kind  string
// }

// // NewResourceController returns a new ResourceController
// func NewResourceController(store store.ResourceStore, kind string) ResourceController {
// 	return ResourceController{store: store, kind: kind}
// }

// // CreateResource ...
// func (c ResourceController) CreateResource(ctx context.Context, resource corev2.Resource) error {
// 	if err := c.store.CreateResource(ctx, resource); err != nil {
// 		switch err := err.(type) {
// 		case *store.ErrAlreadyExists:
// 			return NewErrorf(AlreadyExistsErr)
// 		case *store.ErrNotValid:
// 			return NewErrorf(InvalidArgument)
// 		default:
// 			return NewError(InternalErr, err)
// 		}
// 	}

// 	return nil
// }

// // CreateOrUpdateResource ...
// func (c ResourceController) CreateOrUpdateResource(ctx context.Context, resource corev2.Resource) error {
// 	if err := c.store.CreateOrUpdateResource(ctx, resource); err != nil {
// 		switch err := err.(type) {
// 		case *store.ErrNotValid:
// 			return NewError(InvalidArgument, err)
// 		default:
// 			return NewError(InternalErr, err)
// 		}
// 	}

// 	return nil
// }

// // DeleteResource ...
// func (c ResourceController) DeleteResource(ctx context.Context, name string) error {
// 	if err := c.store.DeleteResource(ctx, c.kind, name); err != nil {
// 		switch err := err.(type) {
// 		case *store.ErrNotFound:
// 			return NewErrorf(NotFound)
// 		default:
// 			return NewError(InternalErr, err)
// 		}
// 	}

// 	return nil
// }

// // GetResource ...
// func (c ResourceController) GetResource(ctx context.Context, name string, resource corev2.Resource) error {
// 	if err := c.store.GetResource(ctx, name, resource); err != nil {
// 		switch err := err.(type) {
// 		case *store.ErrNotFound:
// 			return NewErrorf(NotFound)
// 		default:
// 			return NewError(InternalErr, err)
// 		}
// 	}

// 	return nil
// }
