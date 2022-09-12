package postgres

const isInitializedQuery = `
-- This query determines if the cluster is initialized.
--
SELECT true FROM initialized WHERE initialized = true;
`
const flagAsInitializedQuery = `
-- This query flags the cluster as being initialized.
--
INSERT INTO initialized(initialized) VALUES(true);
`
