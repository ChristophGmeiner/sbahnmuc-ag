#!/bin/bash

sqlite3 telemetry.db "DELETE FROM delay_records;"
sqlite3 telemetry.db "DELETE FROM station_request_logs;"
sqlite3 telemetry.db "VACUUM;"

bq query --nouse_legacy_sql "TRUNCATE TABLE \`$BQ_PROJECT_ID.$BQ_DATASET_ID.delay_records\`;"
bq query --nouse_legacy_sql "TRUNCATE TABLE \`$BQ_PROJECT_ID.$BQ_DATASET_ID.station_request_logs\`;"
