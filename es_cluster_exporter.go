package main

import (
	"flag"
	"fmt"
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
	namespace = "es_cluster"
	//namespace = "elasticsearch_cluster"
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
		"indices_search_fetch_current": &VecInfo{
			help:   "Number of query fetches currently running",
			labels: []string{"cluster", "node"},
		},
		"indices_search_open_contexts": &VecInfo{
			help:   "Number of active searches",
			labels: []string{"cluster", "node"},
		},
		"indices_search_query_current": &VecInfo{
			help:   "Number of currently active queries",
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
		"status": &VecInfo{
			help:   "Cluster status (0=green, 1=yellow, 2=red)",
			labels: []string{"cluster"},
		},
		"nodes_total": &VecInfo{
			help:   "Total number of nodes",
			labels: []string{"cluster"},
		},
		"nodes_data": &VecInfo{
			help:   "Number of data nodes",
			labels: []string{"cluster"},
		},
		"shards_unassigned": &VecInfo{
			help:   "Number of unassigned shards",
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

// getNodeName returns the name of the node that we're monitoring
func (e *Exporter) getNodeName() (name string, err error) {
	var nidx NodeIndex

	resp, err := e.client.Get(e.URI)
	if err != nil {
		log.Println("Error while querying Elasticsearch:", err)
		return
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&nidx); err != nil {
		log.Println("Failed to unmarshal JSON into struct:", err)
		return
	}

	name = nidx.NodeName

	return
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

func (e *Exporter) CollectClusterIndexHealth() {
	var uri string
	var stats ClusterHealthResponse

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

	if err = json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Println("Failed to unmarshal JSON into struct:", err)
		return
	}

	var statusMapping = map[string]float64{"green": 0, "yellow": 1, "red": 2}

	e.gauges["status"].WithLabelValues(stats.ClusterName).Set(statusMapping[stats.Status])
	e.gauges["nodes_total"].WithLabelValues(stats.ClusterName).Set(float64(stats.NumberOfNodes))
	e.gauges["nodes_data"].WithLabelValues(stats.ClusterName).Set(float64(stats.NumberOfDataNodes))
	e.gauges["shards_unassigned"].WithLabelValues(stats.ClusterName).Set(float64(stats.UnassignedShards))

	for indexName, indexStats := range stats.Indices {
		e.gauges["index_status"].WithLabelValues(stats.ClusterName, indexName).Set(statusMapping[indexStats.Status])
		e.gauges["index_shards_active_primary"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.ActivePrimaryShards))
		e.gauges["index_shards_active"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.ActiveShards))
		e.gauges["index_shards_relocating"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.RelocatingShards))
		e.gauges["index_shards_initializing"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.InitializingShards))
		e.gauges["index_shards_unassigned"].WithLabelValues(stats.ClusterName, indexName).Set(float64(indexStats.UnassignedShards))
	}
}

func (e *Exporter) CollectClusterHealth() {
	var uri string
	var allStats NodeStatsResponse

	nodeName, err := e.getNodeName()
	if err != nil {
		e.up.Set(0)
		log.Println("Error while querying Elasticsearch:", err)
		return
	}

	if e.allNodes {
		uri = e.URI + "/_nodes/stats"
	} else {
		uri = e.URI + fmt.Sprintf("/_nodes/%s/stats", nodeName)
	}

	resp, err := e.client.Get(uri)
	if err != nil {
		e.up.Set(0)
		log.Println("Error while querying Elasticsearch:", err)
		return
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&allStats); err != nil {
		log.Println("Failed to unmarshal JSON into struct:", err)
		return
	}

	// If we aren't polling all nodes, make sure we only got one response.
	if !e.allNodes && len(allStats.Nodes) != 1 {
		log.Println("Unexpected number of nodes returned.")
	}

	for _, stats := range allStats.Nodes {
		// GC Stats
		for collector, gcstats := range stats.JVM.GC.Collectors {
			e.counters["jvm_gc_collection_seconds_count"].WithLabelValues(allStats.ClusterName, stats.Name, collector).Set(float64(gcstats.CollectionCount))
			e.counters["jvm_gc_collection_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Name, collector).Set(float64(gcstats.CollectionTime / 1000))
		}

		// Breaker stats
		for breaker, bstats := range stats.Breakers {
			e.gauges["breakers_estimated_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name, breaker).Set(float64(bstats.EstimatedSize))
			e.gauges["breakers_limit_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name, breaker).Set(float64(bstats.LimitSize))
		}

		// Thread Pool stats
		for pool, pstats := range stats.ThreadPool {
			e.counters["thread_pool_completed_count"].WithLabelValues(allStats.ClusterName, stats.Name, pool).Set(float64(pstats.Completed))
			e.counters["thread_pool_rejected_count"].WithLabelValues(allStats.ClusterName, stats.Name, pool).Set(float64(pstats.Rejected))

			e.gauges["thread_pool_active_count"].WithLabelValues(allStats.ClusterName, stats.Name, pool).Set(float64(pstats.Active))
			e.gauges["thread_pool_threads_count"].WithLabelValues(allStats.ClusterName, stats.Name, pool).Set(float64(pstats.Threads))
			e.gauges["thread_pool_largest_count"].WithLabelValues(allStats.ClusterName, stats.Name, pool).Set(float64(pstats.Largest))
			e.gauges["thread_pool_queue_count"].WithLabelValues(allStats.ClusterName, stats.Name, pool).Set(float64(pstats.Queue))
		}

		// JVM Memory Stats
		e.gauges["jvm_memory_committed_bytes"].WithLabelValues(allStats.ClusterName, stats.Name, "heap").Set(float64(stats.JVM.Mem.HeapCommitted))
		e.gauges["jvm_memory_used_bytes"].WithLabelValues(allStats.ClusterName, stats.Name, "heap").Set(float64(stats.JVM.Mem.HeapUsed))
		e.gauges["jvm_memory_max_bytes"].WithLabelValues(allStats.ClusterName, stats.Name, "heap").Set(float64(stats.JVM.Mem.HeapMax))
		e.gauges["jvm_memory_committed_bytes"].WithLabelValues(allStats.ClusterName, stats.Name, "non-heap").Set(float64(stats.JVM.Mem.NonHeapCommitted))
		e.gauges["jvm_memory_used_bytes"].WithLabelValues(allStats.ClusterName, stats.Name, "non-heap").Set(float64(stats.JVM.Mem.NonHeapUsed))

		// Indices Stats
		e.gauges["indices_search_query_current"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Search.QueryCurrent))
		e.gauges["indices_search_open_contexts"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Search.OpenContext))
		e.gauges["indices_search_fetch_current"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Search.FetchCurrent))

		e.gauges["indices_fielddata_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FieldData.MemorySize))
		e.gauges["indices_fielddata_cache_size"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FieldData.CacheSize))
		e.counters["indices_fielddata_evictions"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FieldData.Evictions))
		e.counters["indices_fielddata_hit_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FieldData.HitCount))
		e.counters["indices_fielddata_total_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FieldData.TotalCount))

		e.gauges["indices_filter_cache_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FilterCache.MemorySize))
		e.gauges["indices_filter_cache_cache_size"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FilterCache.CacheSize))
		e.counters["indices_filter_cache_evictions"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FilterCache.Evictions))
		e.counters["indices_filter_cache_hit_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FilterCache.HitCount))
		e.counters["indices_filter_cache_total_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.FilterCache.TotalCount))

		e.gauges["indices_query_cache_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.QueryCache.MemorySize))
		e.gauges["indices_query_cache_cache_size"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.QueryCache.CacheSize))
		e.counters["indices_query_cache_evictions"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.QueryCache.Evictions))
		e.counters["indices_query_cache_hit_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.QueryCache.HitCount))
		e.counters["indices_query_cache_total_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.QueryCache.TotalCount))

		e.gauges["indices_request_cache_memory_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.RequestCache.MemorySize))
		e.gauges["indices_request_cache_cache_size"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.RequestCache.CacheSize))
		e.counters["indices_request_cache_evictions"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.RequestCache.Evictions))
		e.counters["indices_request_cache_hit_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.RequestCache.HitCount))
		e.counters["indices_request_cache_total_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.RequestCache.TotalCount))

		e.counters["indices_translog_operations"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Translog.Operations))
		e.counters["indices_translog_size_in_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Translog.Size))

		e.counters["indices_get_time_ms"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Get.Time))
		e.counters["indices_get_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Get.Total))

		e.counters["indices_get_missing_time_ms"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Get.MissingTime))
		e.counters["indices_get_missing_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Get.MissingTotal))

		e.counters["indices_get_exists_time_ms"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Get.ExistsTime))
		e.counters["indices_get_exists_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Get.ExistsTotal))

		e.counters["indices_search_query_time_ms"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Search.QueryTime))
		e.counters["indices_search_query_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Search.QueryTotal))

		e.counters["indices_search_fetch_time_ms"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Search.FetchTime))
		e.counters["indices_search_fetch_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Search.FetchTotal))

		e.gauges["indices_docs"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Docs.Count))
		e.gauges["indices_docs_deleted"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Docs.Deleted))

		e.gauges["indices_segments_memory_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Segments.Memory))
		e.gauges["indices_segments_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Segments.Count))

		e.gauges["indices_store_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Store.Size))
		e.counters["indices_store_throttle_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Store.ThrottleTime))

		e.counters["indices_flush_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Flush.Total))
		e.counters["indices_flush_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Flush.Time))

		e.counters["indices_indexing_index_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Indexing.IndexTime))
		e.counters["indices_indexing_index_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Indexing.IndexTotal))

		e.counters["indices_indexing_delete_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Indexing.DeleteTime))
		e.counters["indices_indexing_delete_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Indexing.DeleteTotal))

		e.counters["indices_merges_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Merges.TotalTime))
		e.counters["indices_merges_size_bytes_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Merges.TotalSize))
		e.counters["indices_merges_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Merges.Total))
		e.counters["indices_merges_docs_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Merges.TotalDocs))

		e.counters["indices_refresh_time_ms_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Refresh.TotalTime))
		e.counters["indices_refresh_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Indices.Refresh.Total))

		// Transport Stats
		e.counters["transport_rx_packets_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Transport.RxCount))
		e.counters["transport_rx_size_bytes_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Transport.RxSize))
		e.counters["transport_tx_packets_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Transport.TxCount))
		e.counters["transport_tx_size_bytes_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Transport.TxSize))

		// Process Stats
		e.gauges["process_cpu_percent"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Process.CPU.Percent))
		e.gauges["process_mem_resident_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Process.Memory.Resident))
		e.gauges["process_mem_share_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Process.Memory.Share))
		e.gauges["process_mem_virtual_size_bytes"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Process.Memory.TotalVirtual))
		e.gauges["process_open_files_count"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Process.OpenFD))

		e.counters["process_cpu_time_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Name, "total").Set(float64(stats.Process.CPU.Total / 1000))
		e.counters["process_cpu_time_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Name, "sys").Set(float64(stats.Process.CPU.Sys / 1000))
		e.counters["process_cpu_time_seconds_sum"].WithLabelValues(allStats.ClusterName, stats.Name, "user").Set(float64(stats.Process.CPU.User / 1000))

		e.gauges["os_mem_used_percent"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.OS.Mem.UsedPercent))
		// HTTP Stats
		e.counters["http_open_total"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Http.TotalOpen))
		e.gauges["http_open"].WithLabelValues(allStats.ClusterName, stats.Name).Set(float64(stats.Http.CurrentOpen))
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
	e.CollectClusterIndexHealth()
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
		listenAddress = flag.String("web.listen-address", ":9114", "Address to listen on for web interface and telemetry.")
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
