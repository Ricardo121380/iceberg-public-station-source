#!/bin/bash
set -euo pipefail
root=/tmp/iceberg-abuse-review-matrix-20260907
mkdir -p "$root"
chmod 700 "$root"
# Isolated test-only names and loopback ports; never use production DB credentials.
cleanup() {
 sudo docker rm -fv iceberg-abuse-review-test-mysql iceberg-abuse-review-test-postgres >/dev/null 2>&1 || true
}
trap cleanup EXIT
if sudo docker inspect iceberg-abuse-review-test-mysql >/dev/null 2>&1 || sudo docker inspect iceberg-abuse-review-test-postgres >/dev/null 2>&1; then
 trap - EXIT
 echo 'Test container names already exist; refusing reuse' >&2; exit 1
fi
sudo docker run -d --name iceberg-abuse-review-test-mysql --memory 768m --cpus 1 -e MYSQL_ROOT_PASSWORD=isolated-review-test -e MYSQL_DATABASE=abuse_review_test -p 127.0.0.1:23316:3306 mysql:8.0 >/dev/null
sudo docker run -d --name iceberg-abuse-review-test-postgres --memory 256m --cpus 1 -e POSTGRES_PASSWORD=isolated-review-test -e POSTGRES_DB=abuse_review_test -p 127.0.0.1:25436:5432 postgres:17-bookworm >/dev/null
for i in $(seq 1 50); do
 if sudo docker exec iceberg-abuse-review-test-mysql mysql -h127.0.0.1 -uroot -pisolated-review-test -N -e "SELECT 1" >/dev/null 2>&1 && sudo docker exec iceberg-abuse-review-test-postgres pg_isready -U postgres -d abuse_review_test >/dev/null; then break; fi
 sleep 1
done
sudo docker exec iceberg-abuse-review-test-mysql mysql -uroot -pisolated-review-test -N -e 'SELECT VERSION()' 2>/dev/null
sudo docker exec iceberg-abuse-review-test-postgres psql -U postgres -d abuse_review_test -At -c 'SELECT version()'
ABUSE_REVIEW_TEST_MYSQL_DSN='root:isolated-review-test@tcp(127.0.0.1:23316)/abuse_review_test?charset=utf8mb4&parseTime=True&loc=Local' \
ABUSE_REVIEW_TEST_POSTGRES_DSN='host=127.0.0.1 port=25436 user=postgres password=isolated-review-test dbname=abuse_review_test sslmode=disable' \
 /tmp/iceberg-abuse-review-model.test -test.run '^TestAbuseReviewMigrationMatrix$' -test.v | tee "$root/results.txt"
