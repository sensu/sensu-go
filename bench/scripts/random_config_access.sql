\set namespace_num random(0, 16)
\set type_num random(0, 32)
\set resource_num random(0, 1024)

SELECT * FROM configuration
WHERE api_version = 'core/v2'
    AND type = CONCAT('type-', :type_num)
    AND namespace = CONCAT('ns-', :namespace_num)
    AND name = CONCAT('rn-', :resource_num);

