-- https://www.postgresql.org/docs/14/functions-json.html

SELECT routine_name
FROM information_schema.routines
WHERE routine_type = 'FUNCTION' order by routine_name;

-- SELECT jsonb_path_query_first_item( _jsonb->'value', '$.score[*] ? (@ == 24 )' ) FROM "values"."values" WHERE (_jsonb->'name' = to_jsonb('array-embedded'::text));
CREATE OR REPLACE FUNCTION  jsonb_path_query_first_item(arrValue jsonb, condition jsonpath) RETURNS jsonb AS $$
DECLARE
    elem RECORD;
BEGIN
    -- for not an arrays
    IF jsonb_array_length(arrValue) = 0 THEN
        RETURN null;
    END IF;

    FOR elem IN ( SELECT jsonb_array_elements(arrValue) val )
    LOOP
        raise notice 'item %', elem.val;
        IF elem.val @? condition THEN
            RETURN elem.val;
        END IF;
    END LOOP;

    RETURN null;

END;
$$ LANGUAGE plpgsql;


SELECT
 CASE
    WHEN  jsonb_typeof(_jsonb->'value') != 'array' THEN null
    ELSE
        (
            SELECT tempTable.value result
            FROM jsonb_array_elements(_jsonb->'value') tempTable
            WHERE tempTable.value @? '$.score[*] ? (@ == 24 )'
            LIMIT 1
        )
 END val
FROM "values"."values"
WHERE (_jsonb->'name' = to_jsonb('array-embedded'::text));
