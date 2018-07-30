CREATE DATABASE IF NOT EXISTS k8s;
CREATE TABLE IF NOT EXISTS k8s.metrics(date Date DEFAULT toDate(0),name String,cluster String,tags Array(String),val Float64,ts DateTime,updated DateTime DEFAULT now())ENGINE = ReplicatedMergeTree('/clickhouse/tables/1/k8s.metrics','dev242', date, (name, cluster, tags, ts), 8192);
CREATE TABLE IF NOT EXISTS k8s.metrics_all(date Date DEFAULT toDate(0),name String,cluster String,tags Array(String),val Float64,ts DateTime,updated DateTime DEFAULT now()) ENGINE = Distributed(prometheus_ck_cluster, k8s, metrics, sipHash64(name));
