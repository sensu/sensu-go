package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	utilbytes "github.com/sensu/sensu-go/util/bytes"
)

const JWTName = "jwt"

type JWT struct {
	Store storev2.Interface
}

func (j JWT) GetSecret(ctx context.Context) ([]byte, error) {
	keystore := storev2.NewGenericStore[*corev3.SymmetricKey](j.Store)
	key, err := keystore.Get(ctx, storev2.ID{Name: JWTName})
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			randomBytes, err := utilbytes.Random(32)
			if err != nil {
				return nil, fmt.Errorf("error creating jwt secret: %s", err)
			}
			key := &corev3.SymmetricKey{Metadata: corev2.NewObjectMetaP(JWTName, ""), Value: randomBytes}
			if err := keystore.CreateIfNotExists(ctx, key); err != nil {
				if _, ok := err.(*store.ErrAlreadyExists); ok {
					return j.GetSecret(ctx)
				}
				return nil, err
			}
			return key.Value, nil
		} else {
			return nil, err
		}
	}
	return key.Value, nil
}
