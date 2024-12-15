package main

import (
	"os"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger := CreateLogger("info")

	cluster := CreateCluster(gocql.All, "catalog", "scylla-node1", "scylla-node2", "scylla-node3")
	session, err := gocql.NewSession(*cluster)
	if err != nil {
		logger.Fatal("unable to connect to scylla", zap.Error(err))
	}
	defer session.Close()

	SelectQuery(session, logger)
	insertQuery(session, logger)
	SelectQuery(session, logger)
	deleteQuery(session, logger)
	SelectQuery(session, logger)
}

func CreateLogger(level string) *zap.Logger {
	lvl := zap.NewAtomicLevel()
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl.SetLevel(zap.InfoLevel)
	}
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		lvl,
	))
	return logger
}

func CreateCluster(consistency gocql.Consistency, keyspace string, hosts ...string) *gocql.ClusterConfig {
	retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
		Min:        time.Second,
		Max:        10 * time.Second,
		NumRetries: 5,
	}
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Timeout = 5 * time.Second
	cluster.RetryPolicy = retryPolicy
	cluster.Consistency = consistency
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	return cluster
}

func SelectQuery(session *gocql.Session, logger *zap.Logger) {
	logger.Info("Displaying Results:")
	q := session.Query("SELECT first_name,last_name,address,picture_location FROM mutant_data")
	var firstName, lastName, address, pictureLocation string
	it := q.Iter()
	defer func() {
		if err := it.Close(); err != nil {
			logger.Warn("select catalog.mutant", zap.Error(err))
		}
	}()
	for it.Scan(&firstName, &lastName, &address, &pictureLocation) {
		logger.Info("\t" + firstName + " " + lastName + ", " + address + ", " + pictureLocation)
	}
}

func insertQuery(session *gocql.Session, logger *zap.Logger) {
	logger.Info("Inserting Mike")
	if err := session.Query("INSERT INTO mutant_data (first_name,last_name,address,picture_location) VALUES ('Mike','Tyson','1515 Main St', 'http://www.facebook.com/mtyson')").Exec(); err != nil {
		logger.Error("insert catalog.mutant_data", zap.Error(err))
	}
}

func deleteQuery(session *gocql.Session, logger *zap.Logger) {
	logger.Info("Deleting Mike")
	if err := session.Query("DELETE FROM mutant_data WHERE first_name = 'Mike' and last_name = 'Tyson'").Exec(); err != nil {
		logger.Error("delete catalog.mutant_data", zap.Error(err))
	}
}
