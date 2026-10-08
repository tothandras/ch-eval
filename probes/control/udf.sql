-- probe: SQL user-defined functions (SM-03; executable UDFs need server config and are assessed)
-- cleanup: DROP FUNCTION IF EXISTS ch_eval_probe_udf
-- expect: value:42
CREATE FUNCTION ch_eval_probe_udf AS (x) -> x + 1;
SELECT ch_eval_probe_udf(41)
