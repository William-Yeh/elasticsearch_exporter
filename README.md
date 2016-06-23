# Elasticsearch Cluster Exporter

Export Elasticsearch cluster-level health to Prometheus.

To run it:

```bash
go build
./es_cluster_exporter [flags]
```

### Flags

```bash
./es_cluster_exporter --help
```

* __`es.uri`:__ Address (host and port) of the Elasticsearch node we should
    connect to. This could be a local node (`localhost:8500`, for instance), or
    the address of a remote Elasticsearch server.
* __`es.all`:__ If true, query stats for all nodes in the cluster,
    rather than just the node we connect to.
* __`es.timeout`:__ Timeout for trying to get stats from Elasticsearch. (ex: 20s)
* __`web.listen-address`:__ Address to listen on for web interface and telemetry.
* __`web.telemetry-path`:__ Path under which to expose metrics.

__NOTE:__ We support pulling stats for all nodes at once, but in production
this is unlikely to be the way you actually want to run the system. It is much
better to run an exporter on each Elasticsearch node to remove a single point
of failure and improve the connection between operation and reporting.



## History

The first version of this repo was forked from [f1yegor/elasticsearch_exporter](https://github.com/f1yegor/elasticsearch_exporter), which in turn forked from the *origin* [ewr/elasticsearch_exporter](https://github.com/ewr/elasticsearch_exporter).

The origin [ewr/elasticsearch_exporter](https://github.com/ewr/elasticsearch_exporter) has basic per-node metrics, but the maintainer hasn't adopted newer PRs for some period of time. Some PRs and forks (e.g., [f1yegor/elasticsearch_exporter](https://github.com/f1yegor/elasticsearch_exporter), [hudashot/elasticsearch_exporter](hudashot/elasticsearch_exporter), [issue #5](https://github.com/ewr/elasticsearch_exporter/issues/5), and [PR #13](https://github.com/ewr/elasticsearch_exporter/pull/13)) tries to add cluster-level health status for Elasticsearch, but none is dominant.

Among them, I like the [f1yegor/elasticsearch_exporter](https://github.com/f1yegor/elasticsearch_exporter) folk, but I have no enough time to merge this to my colleague's folk [nodestory/elasticsearch_exporter](https://github.com/nodestory/elasticsearch_exporter). Therefore, I decide to extract the *cluster-related part* of [f1yegor/elasticsearch_exporter](https://github.com/f1yegor/elasticsearch_exporter), and rename it so that it can be used together with other elasticsearch_exporter variants.