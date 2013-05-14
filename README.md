## hutil

hutil is a package of utilities for Go `net/http` servers.

## apachelog

## fileserver

## handler caching

(Cache the response according to the headers.)

## Goals for hutil/metrics:

1. For discrete dimensions of requests (response status, path [bucketed -- this is hard], ...)
  * Keep 1 day of 10s buckets of counts in a ring buffer
  * Show counts for last 5 minutes, hour, day
  * Show plots for last 5 minutes (by 10s), hour (by 30s), day (by 10min)
2. For continuous dimensions of requests (response latency, response size, ...)
  * Implement time-bounded sliding windows from Xu/Lin/Zhou, and use it to show mean, median, 90th, 99th
    quantiles for the last 5 minutes, hour, day
  * Use github.com/bmizerany/perks/quantile to keep (10s, 30s, and 10min) (mean, median, 90th, 99th
    percentile) and store in a ring buffer. Show plots for last 5 minutes (by 10s), hour (by 30s), and day
    (by 10min).
  * Keep a ring buffer with the top 5 requests by $METRIC for (10s, 2m, 30min) buckets and use it to
    calculate the top 5 overall responses for the last (5m, 1h, day). (It is approximate because the
    cutoff on the back end of the time period isn't exact.)
