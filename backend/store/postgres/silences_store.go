package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

type SilenceStore struct {
	db DBI
}

func NewSilenceStore(db *pgxpool.Pool) *SilenceStore {
	return &SilenceStore{db: db}
}

const deleteSilencesQuery = `
WITH ns AS (
	SELECT id FROM namespaces where name = $1 AND deleted_at IS NULL
)
DELETE FROM silences WHERE namespace = (select id from ns) AND name = ANY($2);
`

func (s *SilenceStore) DeleteSilences(ctx context.Context, namespace string, names []string) error {
	_, err := s.db.Exec(ctx, deleteSilencesQuery, namespace, names)
	return err
}

const getSilencesQuery = `
SELECT
	namespaces.name,
	silences.name,
	silences.labels,
	silences.annotations,
	silences.subscription,
	silences.check_name,
	silences.reason,
	silences.expire_on_resolve,
	silences.begin,
	silences.expire_at
FROM
	silences
	JOIN namespaces ON silences.namespace = namespaces.id
	WHERE namespaces.name = $1 OR $1 = '';
`

type scanFunc func(...interface{}) error

func readSilence(sf scanFunc) (*corev2.Silenced, error) {
	var result corev2.Silenced
	var labels []byte
	var annotations []byte
	err := sf(
		&result.ObjectMeta.Namespace,
		&result.ObjectMeta.Name,
		&labels,
		&annotations,
		&result.Subscription,
		&result.Check,
		&result.Reason,
		&result.ExpireOnResolve,
		&result.Begin,
		&result.ExpireAt,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *SilenceStore) GetSilences(ctx context.Context, namespace string) ([]*corev2.Silenced, error) {
	rows, err := s.db.Query(ctx, getSilencesQuery, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*corev2.Silenced{}
	for rows.Next() {
		silenced, err := readSilence(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, silenced)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

const getSilencesByCheckQuery = `
SELECT
	namespaces.name,
	silences.name,
	silences.labels,
	silences.annotations,
	silences.subscription,
	silences.check_name,
	silences.reason,
	silences.expire_on_resolve,
	silences.begin,
	silences.expire_at
FROM
	silences, namespaces
WHERE
	silences.namespace = namespaces.id
	AND namespaces.name = $1
	AND silences.check_name = $2;
`

func (s *SilenceStore) GetSilencesByCheck(ctx context.Context, namespace, check string) ([]*corev2.Silenced, error) {
	rows, err := s.db.Query(ctx, getSilencesByCheckQuery, namespace, check)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*corev2.Silenced{}
	for rows.Next() {
		silenced, err := readSilence(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, silenced)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

const getSilencesBySubscriptionQuery = `
WITH subscriptions AS (
	SELECT UNNEST($2::text[]) AS subscription
)
SELECT
	namespaces.name,
	silences.name,
	silences.labels,
	silences.annotations,
	silences.subscription,
	silences.check_name,
	silences.reason,
	silences.expire_on_resolve,
	silences.begin,
	silences.expire_at
FROM
	silences, namespaces, subscriptions
WHERE
	silences.namespace = namespaces.id
	AND namespaces.name = $1
	AND silences.subscription = subscriptions.subscription;
`

func (s *SilenceStore) GetSilencesBySubscription(ctx context.Context, namespace string, subscriptions []string) ([]*corev2.Silenced, error) {
	rows, err := s.db.Query(ctx, getSilencesBySubscriptionQuery, namespace, subscriptions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*corev2.Silenced{}
	for rows.Next() {
		silenced, err := readSilence(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, silenced)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

const getSilenceByNameQuery = `
SELECT
	namespaces.name,
	silences.name,
	silences.labels,
	silences.annotations,
	silences.subscription,
	silences.check_name,
	silences.reason,
	silences.expire_on_resolve,
	silences.begin,
	silences.expire_at
FROM
	silences, namespaces
WHERE
	silences.namespace = namespaces.id
	AND namespaces.name = $1
	AND silences.name = $2
LIMIT 1;
`

func (s *SilenceStore) GetSilenceByName(ctx context.Context, namespace, name string) (*corev2.Silenced, error) {
	row := s.db.QueryRow(ctx, getSilenceByNameQuery, namespace, name)
	return readSilence(row.Scan)
}

const updateSilencesQuery = `
WITH ns AS (
	SELECT id AS id
	FROM namespaces
	WHERE namespaces.name = $1
	LIMIT 1
)
INSERT INTO silences (
	namespace,
	name,
	labels,
	annotations,
	subscription,
	check_name,
	reason,
	expire_on_resolve,
	begin,
	expire_at
) SELECT ns.id, $2, $3, $4, $5, $6, $7, $8, $9, $10 FROM ns
ON CONFLICT ( namespace, name ) DO UPDATE
SET (
	labels,
	annotations,
	subscription,
	check_name,
	reason,
	expire_on_resolve,
	begin,
	expire_at
) = (
	$3,
	$4,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10
);
`

func (s *SilenceStore) UpdateSilence(ctx context.Context, si *corev2.Silenced) error {
	labels, _ := json.Marshal(si.Labels)
	annos, _ := json.Marshal(si.Annotations)
	result, err := s.db.Exec(
		ctx,
		updateSilencesQuery,
		si.Namespace,
		si.Name,
		labels,
		annos,
		si.Subscription,
		si.Check,
		si.Reason,
		si.ExpireOnResolve,
		si.Begin,
		si.ExpireAt,
	)
	if err != nil {
		return err
	}
	if n := result.RowsAffected(); n == 0 {
		return &store.ErrNamespaceMissing{
			Namespace: si.Namespace,
		}
	}
	return nil
}

const getSilencesByNameQuery = `
WITH names AS (
	SELECT UNNEST($2::text[]) AS name
)
SELECT
	namespaces.name,
	silences.name,
	silences.labels,
	silences.annotations,
	silences.subscription,
	silences.check_name,
	silences.reason,
	silences.expire_on_resolve,
	silences.begin,
	silences.expire_at
FROM
	silences, namespaces, names
WHERE
	silences.namespace = namespaces.id
	AND namespaces.name = $1
	AND silences.name = names.name;
`

func (s *SilenceStore) GetSilencesByName(ctx context.Context, namespace string, names []string) ([]*corev2.Silenced, error) {
	rows, err := s.db.Query(ctx, getSilencesByNameQuery, namespace, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*corev2.Silenced{}
	for rows.Next() {
		silenced, err := readSilence(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, silenced)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
