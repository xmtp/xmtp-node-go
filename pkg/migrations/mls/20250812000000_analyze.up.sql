CREATE OR REPLACE FUNCTION vacuum_analyze_table(tbl_name text)
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
EXECUTE format(
        'SET LOCAL statement_timeout = 0; VACUUM (ANALYZE, VERBOSE, PARALLEL 4) %I',
        tbl_name
        );
END;
$$;