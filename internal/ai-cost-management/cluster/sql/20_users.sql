-- Data-team provisioned users (D-14). Local test passwords only.
CREATE USER IF NOT EXISTS aicm_reader IDENTIFIED WITH plaintext_password BY 'reader' DEFAULT DATABASE ai_cost_management;
GRANT ai_cost_management_reader TO aicm_reader;
ALTER USER aicm_reader DEFAULT ROLE ai_cost_management_reader;
CREATE USER IF NOT EXISTS aicm_writer IDENTIFIED WITH plaintext_password BY 'writer' DEFAULT DATABASE ai_cost_management;
GRANT ai_cost_management_writer TO aicm_writer;
ALTER USER aicm_writer DEFAULT ROLE ai_cost_management_writer;
