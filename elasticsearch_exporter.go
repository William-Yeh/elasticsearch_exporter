package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "elasticsearch_cluster"
	//namespace = "elasticsearch"
)

type VecInfo struct {
	help   string
	labels []string
}

var (
	counterMetrics = map[string]*VecInfo{
		"indices_fielddata_evictions": &VecInfo{
			help:   "Evictions from field data",
			labels: []string{"cluster", "node"},
		},
		"indices_filter_cache_evictions": &VecInfo{
			help:   "Evictions from filter cache",
			labels: []string{"cluster", "node"},
		},
		"indices_query_cache_evictions": &VecInfo{
			help:   "Evictions from query cache",
			labels: []string{"cluster", "node"},
		},
		"indices_request_cache_evictions": &VecInfo{
			help:   "Evictions from request cache",
			labels: []string{"cluster", "node"},
		},
		"indices_fielddata_hit_count": &VecInfo{
			help:   "Hit count of field data",
			labels: []string{"cluster", "node"},
		},
		"indices_filter_cache_hit_count": &VecInfo{
			help:   "Hit count of filter cache",
			labels: []string{"cluster", "node"},
		},
		"indices_query_cache_hit_count": &VecInfo{
			help:   "Hit count of query cache",
			labels: []string{"cluster", "node"},
		},
		"indices_request_cache_hit_count": &VecInfo{
			help:   "Hit count of request cache",
			labels: []string{"cluster", "node"},
		},
		"indices_fielddata_total_count": &VecInfo{
			help:   "Total request count of field data",
			labels: []string{"cluster", "node"},
		},
		"indices_filter_cache_total_count": &VecInfo{
			help:   "Total request count of filter cache",
			labels: []string{"cluster", "node"},
		},
		"indices_query_cache_total_count": &VecInfo{
			help:   "Total request count of query cache",
			labels: []string{"cluster", "node"},
		},
		"indices_request_cache_total_count": &VecInfo{
			help:   "Total request count of request cache",
			labels: []string{"cluster", "node"},
		},
		"indices_flush_total": &VecInfo{
			help:   "Total flushes",
			labels: []string{"cluster", "node"},
		},
		"indices_flush_time_ms_total": &VecInfo{
			help:   "Cumulative flush time in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"transport_rx_packets_total": &VecInfo{
			help:   "Count of packets received",
			labels: []string{"cluster", "node"},
		},
		"transport_rx_size_bytes_total": &VecInfo{
			help:   "Total number of bytes received",
			labels: []string{"cluster", "node"},
		},
		"transport_tx_packets_total": &VecInfo{
			help:   "Count of packets sent",
			labels: []string{"cluster", "node"},
		},
		"transport_tx_size_bytes_total": &VecInfo{
			help:   "Total number of bytes sent",
			labels: []string{"cluster", "node"},
		},
		"indices_store_throttle_time_ms_total": &VecInfo{
			help:   "Throttle time for index store in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_indexing_index_total": &VecInfo{
			help:   "Total index calls",
			labels: []string{"cluster", "node"},
		},
		"indices_indexing_index_time_ms_total": &VecInfo{
			help:   "Cumulative index time in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_merges_total": &VecInfo{
			help:   "Total merges",
			labels: []string{"cluster", "node"},
		},
		"indices_merges_docs_total": &VecInfo{
			help:   "Cumulative docs merged",
			labels: []string{"cluster", "node"},
		},
		"indices_merges_size_bytes_total": &VecInfo{
			help:   "Total merge size in bytes",
			labels: []string{"cluster", "node"},
		},
		"indices_merges_time_ms_total": &VecInfo{
			help:   "Total time spent merging in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_refresh_total": &VecInfo{
			help:   "Total refreshes",
			labels: []string{"cluster", "node"},
		},
		"indices_refresh_time_ms_total": &VecInfo{
			help:   "Total time spent refreshing",
			labels: []string{"cluster", "node"},
		},
		"indices_indexing_delete_total": &VecInfo{
			help:   "Total indexing deletes",
			labels: []string{"cluster", "node"},
		},
		"indices_indexing_delete_time_ms_total": &VecInfo{
			help:   "Total time indexing delete in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_get_time_ms": &VecInfo{
			help:   "Total get time in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_get_total": &VecInfo{
			help:   "Total get",
			labels: []string{"cluster", "node"},
		},
		"indices_get_missing_time_ms": &VecInfo{
			help:   "Total time of get missing in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_get_missing_total": &VecInfo{
			help:   "Total get missing",
			labels: []string{"cluster", "node"},
		},
		"indices_get_exists_time_ms": &VecInfo{
			help:   "Total time get exists in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_get_exists_total": &VecInfo{
			help:   "Total get exists operations",
			labels: []string{"cluster", "node"},
		},
		"indices_search_query_time_ms": &VecInfo{
			help:   "Total search query time in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_search_query_total": &VecInfo{
			help:   "Total search queries",
			labels: []string{"cluster", "node"},
		},
		"indices_search_fetch_time_ms": &VecInfo{
			help:   "Total search fetch time in milliseconds",
			labels: []string{"cluster", "node"},
		},
		"indices_search_fetch_total": &VecInfo{
			help:   "Total search fetches",
			labels: []string{"cluster", "node"},
		},
		"indices_translog_operations": &VecInfo{
			help:   "Total translog operations",
			labels: []string{"cluster", "node"},
		},
		"indices_translog_size_in_bytes": &VecInfo{
			help:   "Total translog size in bytes",
			labels: []string{"cluster", "node"},
		},
		"jvm_gc_collection_seconds_count": &VecInfo{
			help:   "Count of JVM GC runs",
			labels: []string{"cluster", "node", "gc"},
		},
		"jvm_gc_collection_seconds_sum": &VecInfo{
			help:   "GC run time in seconds",
			labels: []string{"cluster", "node", "gc"},
		},
		"process_cpu_time_seconds_sum": &VecInfo{
			help:   "Process CPU time in seconds",
			labels: []string{"cluster", "node", "type"},
		},
		"thread_pool_completed_count": &VecInfo{
			help:   "Thread Pool operations completed",
			labels: []string{"cluster", "node", "type"},
		},
		"thread_pool_rejected_count": &VecInfo{
			help:   "Thread Pool operations rejected",
			labels: []string{"cluster", "node", "type"},
		},
		"http_open_total": &VecInfo{
			help:   "Total HTTP connections opened",
			labels: []string{"cluster", "node"},
		},
	}

	gaugeMetrics = map[string]*VecInfo{
		"indices_fielddata_memory_size_bytes": &VecInfo{
			help:   "Field data cache memory usage in bytes",
			labels: []string{"cluster", "node"},
		},
		"indices_filter_cache_memory_size_bytes": &VecInfo{
			help:   "Filter cache memory usage in bytes",
			labels: []string{"cluster", "node"},
		},
		"indices_query_cache_memory_size_bytes": &VecInfo{
			help:   "Query cache memory usage in bytes",
			labels: []string{"cluster", "node"},
		},
		"indices_request_cache_memory_size_bytes": &VecInfo{
			help:   "Request cache memory usage in bytes",
			labels: []string{"cluster", "node"},
		},
		"indices_fielddata_cache_size": &VecInfo{
			help:   "Field data cache size",
			labels: []string{"cluster", "node"},
		},
		"indices_filter_cache_cache_size": &VecInfo{
			help:   "Filter cache size",
			labels: []string{"cluster", "node"},
		},
		"indices_query_cache_cache_size": &VecInfo{
			help:   "Query cache size",
			labels: []string{"cluster", "node"},
		},
		"indices_request_cache_cache_size": &VecInfo{
			help:   "Request cache size",
			labels: []string{"cluster", "node"},
		},
		"indices_docs": &VecInfo{
			help:   "Count of documents on this node",
			labels: []string{"cluster", "node"},
		},
		"indices_docs_deleted": &VecInfo{
			help:   "Count of deleted documents on this node",
			labels: []string{"cluster", "node"},
		},
		"indices_store_size_bytes": &VecInfo{
			help:   "Current size of stored index data in bytes",
			labels: []string{"cluster", "node"},
		},
		"indices_segments_memory_bytes": &VecInfo{
			help:   "Current memory size of segments in bytes",
			labels: []string{"cluster", "node"},
		},
		"indices_segments_count": &VecInfo{
			help:   "Count of index segments on this node",
			labels: []string{"cluster", "node"},
		},
		"process_cpu_percent": &VecInfo{
			help:   "Percent CPU used by process",
			labels: []string{"cluster", "node"},
		},
		"process_mem_resident_size_bytes": &VecInfo{
			help:   "Resident memory in use by process in bytes",
			labels: []string{"cluster", "node"},
		},
		"process_mem_share_size_bytes": &VecInfo{
			help:   "Shared memory in use by process in bytes",
			labels: []string{"cluster", "node"},
		},
		"process_mem_virtual_size_bytes": &VecInfo{
			help:   "Total virtual memory used in bytes",
			labels: []string{"cluster", "node"},
		},
		"process_open_files_count": &VecInfo{
			help:   "Open file descriptors",
			labels: []string{"cluster", "node"},
		},
		"process_max_files_count": &VecInfo{
			help:   "Max file descriptors for process",
			labels: []string{"cluster", "node"},
		},
		"os_mem_used_percent": &VecInfo{
			help:   "Percentage of used memory",
			labels: []string{"cluster", "node"},
		},
		"breakers_estimated_size_bytes": &VecInfo{
			help:   "Estimated size in bytes of breaker",
			labels: []string{"cluster", "node", "breaker"},
		},
		"breakers_limit_size_bytes": &VecInfo{
			help:   "Limit size in bytes for breaker",
			labels: []string{"cluster", "node", "breaker"},
		},
		"breakers_tripped": &VecInfo{
			help:   "Has the breaker been tripped?",
			labels: []string{"breaker"},
		},
		"jvm_memory_committed_bytes": &VecInfo{
			help:   "JVM memory currently committed by area",
			labels: []string{"cluster", "node", "area"},
		},
		"jvm_memory_used_bytes": &VecInfo{
			help:   "JVM memory currently used by area",
			labels: []string{"cluster", "node", "area"},
		},
		"jvm_memory_max_bytes": &VecInfo{
			help:   "JVM memory max",
			labels: []string{"cluster", "node", "area"},
		},
		"thread_pool_active_count": &VecInfo{
			help:   "Thread Pool threads active",
			labels: []string{"cluster", "node", "type"},
		},
		"thread_pool_largest_count": &VecInfo{
			help:   "Thread Pool largest threads count",
			labels: []string{"cluster", "node", "type"},
		},
		"thread_pool_queue_count": &VecInfo{
			help:   "Thread Pool operations queued",
			labels: []string{"cluster", "node", "type"},
		},
		"thread_pool_threads_count": &VecInfo{
			help:   "Thread Pool current threads count",
			labels: []string{"cluster", "node", "type"},
		},
		"cluster_nodes_total": &VecInfo{
			help:   "Total number of nodes",
			labels: []string{"cluster"},
		},
		"cluster_nodes_data": &VecInfo{
			help:   "Number of data nodes",
			labels: []string{"cluster"},
		},
		"index_status": &VecInfo{
			help:   "Index status (0=green, 1=yellow, 2=red)",
			labels: []string{"cluster", "index"},
		},
		"index_shards_active_primary": &VecInfo{
			help:   "Number of active primary shards",
			labels: []string{"cluster", "index"},
		},
		"index_shards_active": &VecInfo{
			help:   "Number of active shards",
			labels: []string{"cluster", "index"},
		},
		"index_shards_relocating": &VecInfo{
			help:   "Number of relocating shards",
			labels: []string{"cluster", "index"},
		},
		"index_shards_initializing": &VecInfo{
			help:   "Number of initializing shards",
			labels: []string{"cluster", "index"},
		},
		"index_shards_unassigned": &VecInfo{
			help:   "Number of unassigned shards",
			labels: []string{"cluster", "index"},
		},
		"http_open": &VecInfo{
			help:   "Current HTTP connections opened",
			labels: []string{"cluster", "node"},
		},
	}
)

// Exporter collects Elasticsearch stats from the given server and exports
// them using the prometheus metrics package.
type Exporter struct {
	URI   string
	mutex sync.RWMutex

	up prometheus.Gauge

	gauges   map[string]*prometheus.GaugeVec
	counters map[string]*prometheus.CounterVec

	allNodes bool

	client *http.Client
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, timeout time.Duration, allNodes bool) *Exporter {
	counters := make(map[string]*prometheus.CounterVec, len(counterMetrics))
	gauges := make(map[string]*prometheus.GaugeVec, len(gaugeMetrics))

	for name, info := range counterMetrics {
		counters[name] = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      info.help,
		}, info.labels)
	}

	for name, info := range gaugeMetrics {
		gauges[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      name,
			Help:      info.help,
		}, info.labels)
	}

	// Init our exporter.
	return &Exporter{
		URI: uri,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the Elasticsearch instance query successful?",
		}),

		counters: counters,
		gauges:   gauges,

		allNodes: allNodes,

		client: &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, timeout)
					if err != nil {
						return nil, err
					}
					if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
						return nil, err
					}
					return c, nil
				},
			},
		},
	}
}

// Describe describes all the metrics ever exported by the elasticsearch
// exporter. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()

	for _, vec := range e.counters {
		vec.Describe(ch)
	}

	for _, vec := range e.gauges {
		vec.Describe(ch)
	}
}

func (e *Exporter) CollectClusterHealth() {
	var uri string
	if e.allNodes {
		uri = e.URI + "/_cluster/health?level=indices"
	} else {
		uri = e.URI + "/_cluster/health?level=indices&local=true"
	}

	resp, err := e.client.Get(uri)
	if err != nil {
		e.up.Set(0)
		log.Println("Error while querying Elasticsearch:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read ES response body:", err)
		e.up.Set(0)
		return
	}

	var stats ClusterHealthResponse
	err = json.Unmarshal(body, &stats)
	if err != nil {
		log.Println("Failed to unmarshal JSON into struct:", err)
		return
	}

	e.gauges["cluster_nodes_total"].WithLabelValues(stats.ClusterName).Set(float64(stats.NumberOfNodes))
	e.gauges["cluster_nodes_data"].WithLabelValues(stats.ClusterName).Set(float64(stats.NumberOfDataNodes))

	var statusMap = map[string]float64{"green": 0, "yellow": 1, "red": 2}
	for indexName, indexStats := range stats.Indices {
		e.gauges["index_status"].WithLabelValues(stats.ClusterName, indexName).Set(statusMap[indexStats.Status])
		e.gauges["index_shards_active_primary"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.ActivePrimaryShards))
		e.gauges["index_shards_active"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.ActiveShards))
		e.gauges["index_shards_relocating"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.RelocatingShards))
		e.gauges["index_shards_initializing"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.InitializingShards))
		e.gauges["index_shards_unassigned"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.UnassignedShards))
	}
}

// Collect fetches the stats from configured elasticsearch location and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	// Reset metrics.
	for _, vec := range e.gauges {
		vec.Reset()
	}

	for _, vec := range e.counters {
		vec.Reset()
	}

	defer func() { ch <- e.up }()

	// This will be reset to 0 if any of the collectors below fail
	e.up.Set(1)

	// Collect metrics
	e.CollectClusterHealth()

	// Report metrics.
	for _, vec := range e.counters {
		vec.Collect(ch)
	}

	for _, vec := range e.gauges {
		vec.Collect(ch)
	}
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9108", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		esURI         = flag.String("es.uri", "http://localhost:9200", "HTTP API address of an Elasticsearch node.")
		esTimeout     = flag.Duration("es.timeout", 5*time.Second, "Timeout for trying to get stats from Elasticsearch.")
		esAllNodes    = flag.Bool("es.all", false, "Export stats for all nodes in the cluster.")
	)
	flag.Parse()

	exporter := NewExporter(*esURI, *esTimeout, *esAllNodes)
	prometheus.MustRegister(exporter)

	log.Println("Starting Server:", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Elasticsearch Cluster Exporter</title></head>
             <body>
             <h1>Elasticsearch Cluster Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
