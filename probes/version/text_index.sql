-- probe: text (full-text) index, GA form (VG-03)
-- expect: value:1
CREATE TABLE {db}.text_probe (s String, INDEX t lower(s) TYPE text(tokenizer = 'splitByNonAlpha')) ENGINE = MergeTree ORDER BY tuple();
INSERT INTO {db}.text_probe VALUES ('Hello World'), ('other text');
SELECT count() FROM {db}.text_probe WHERE hasToken(lower(s), 'world')
