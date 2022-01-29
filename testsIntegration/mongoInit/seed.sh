#! bin/bash
# drop the test db in case this script is run on an existing database
mongo <<-EOJS
use test
db.dropDatabase()
EOJS

echo "SEEDING TEST DATA"
mongoimport -d test -c temporaryLinks < /testData/temporaryLinks.json
mongoimport -d test -c apiKeys < /testData/apiKeys.json