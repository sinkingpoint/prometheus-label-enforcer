# Label Enforcer

This is a simple reverse proxy that sits in front of a Prometheus API, inspecting queries that come in to the API and rejecting any that don't contain filters on the given label keys.

This is useful for things like [Thanos](https://thanos.io), which uses external labels to limit the potential fan out of its queries. By using the label enforcer, you can enforce that every query that comes though contains one of these filters, limiting the fanout by default and making your queries _speedy_. 

## Usage

```
Usage: label-enforcer --labels=LABELS,...

Flags:
  -h, --help                          Show context-sensitive help.
      --listen-address=":4278"        Address to listen on for HTTP requests.
      --backend-url="http://:9090"    URL of the backend to proxy requests to.
      --labels=LABELS,...             Comma-separated list of labels to enforce.
```

e.g.

```
./label-enforcer --labels colo_name,colo_id
```

will block all query requests that don't have a `colo_name=` or `colo_id=` filter.

Note that you can still achieve heavier queries with a regex, e.g. `my_metric{colo_name=~".+"}`, but the label-enforcer makes those heavier queries opt-in, rather than the default.
