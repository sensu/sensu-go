DELETE FROM counters;
DELETE FROM configuration;

-- insert 5k counters
INSERT INTO counters (c) SELECT 0 FROM generate_series(1, 5000);

-- insert 1k resources per type per namespace
-- 16 namesapces * 32 resource types * 1024 resources = ~500k resources total.
DO $$
DECLARE
    num_silenced int := 5000; 
    num_resources int := 1024;
    num_types int := 32;
    num_namespaces int := 16;
    t text := '';
    ns text := '';
    rn text := '';
BEGIN
FOR i_type IN 1..num_types LOOP
    t := (select concat('type-', i_type));
    FOR i_ns IN 1..num_namespaces LOOP
        ns := (select concat('ns-', i_ns % num_namespaces));
        FOR i_r IN 1..num_resources LOOP
            rn := (select concat('rn-', i_r, '-', i_ns, '-', i_type));
            INSERT INTO configuration (api_version, type, namespace, name, labels, annotations, resource)
            VALUES (
                'core/v2',
                t,
                ns,
                rn,
                '{}',
                '',
                '{"key": "value"}'
            );
        END LOOP;
    END LOOP;
END LOOP;
-- insert 5k silences per namesapce
FOR i_ns IN 1..num_namespaces LOOP
    ns := (select concat('ns-', i_ns % num_namespaces));
    FOR i_s IN 1..num_silenced LOOP
        rn := (select concat('silenced-', i_s, '-', i_ns));
        INSERT INTO configuration (api_version, type, namespace, name, labels, annotations, resource)
        VALUES (
            'core/v2',
            'Silenced',
            ns,
            rn,
            '{}',
            '',
            '{"key": "value"}'
        );
        INSERT INTO silenced (begin_at, expire_at, expire_on_resolve, namespace, name)
        VALUES (
            NOW(),
            NOW() + '33 minutes'::interval,
            false,
            ns,
            rn
        );
    END LOOP;
END LOOP;
END;
$$;

