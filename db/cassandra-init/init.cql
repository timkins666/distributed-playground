CREATE KEYSPACE IF NOT EXISTS payments WITH replication = {'class':'SimpleStrategy','replication_factor':1};
CREATE TABLE IF NOT EXISTS payments.txn (id UUID PRIMARY KEY, amount decimal, created timestamp);
