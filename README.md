# `crawl`

`crawl` is an efficient and concurrent tool for crawling and
understanding web sites. It outputs data in a newline-delimited JSON
format suitable for use with BigQuery.

## Motivation

`crawl` is a response to two challenges with popular crawling tools.
First, most tools combine data collection and analysis. This makes
crawling (which is essentially collection) appear to require more
computing resource than it actually does. Second, they constrain
analysis to pre-defined formats. You rely on the folks building the
crawlers to also understand exactly what analysis you need to perform.

The promise of cloud computing is that you can commission the compute
power you need, when you need it.  BigQuery is a magical example of
this in action. Anyone can upload (possibly nested) data and analyze
it in seconds, without having to maintain infrastructure. Analysis
can be done remotely without any loss of control over what questions
can be asked. `crawl` outputs its data in a format compatible with
BigQuery.

The structure of the data is important. With most crawlers that allow
data exports, the result is tabular. You get, for instance, one row
per page in a CSV. This structure is not able to represent the
one-to-many or many-to-many relationships of cross-linking within a
site. `crawl` outputs a single row per page, but that row contains
nested data about every link, hreflang tag, header field... all of its
important structure is preserved and available for later use.

So relying on BigQuery for analysis solves the second problem (of
flexibility). What about the first? Well, it turns out that if you
don't try to analyze the data at all as you're collecting it, you can
be quite efficient. `crawl` maintains the minimum state necessary to
complete the crawl. In practice, a crawl of a 10,000 page site might
use ~30 MB RAM. Crawling 1,000,000 pages might use less than a
gigabyte.

## Installation

Currently you must build `crawl` from source. This will require
Go >1.10.

```sh
go get -u github.com/benjaminestes/crawl/...
```

In a well-configured Go installation, this should fetch and build the
tool. The binary will be put in your $GOBIN directory. Adding $GOBIN
to your $PATH will allow you to call `crawl` without specifying its
location.

## Use

```
USAGE: crawl <command> [-flags] [args]

The following commands are valid:
        help, list, schema, sitemap, spider

help        Print this message.

list        Crawl a list of URLs provided on stdin.

            The -format={(text)|xml} flag determines the expected type.

            Example:
            crawl list config.json <url_list.txt >out.txt
            crawl list -format=xml config.json <sitemap.xml >out.txt

schema      Print a BigQuery-compatible JSON schema to stdout.

            Example:
            crawl schema >schema.json

sitemap     Recursively requests a sitemap or sitemap index from
            a URL provided as argument.

            Example:
            crawl sitemap http://www.example.com/sitemap.xml >out.txt

spider      Crawl from the URLs specific in the configuration file.

            Example:
            crawl spider config.json >out.txt
```

## Configuration

The repository includes an example `config.json` file. This lists all
of the available options with reasonable default values. In
particular, you should think about these options:

- `From`: An array of fully-qualified URLs from which you want to
    start crawling. If you are crawling from the home page of a site,
    this list will have one item in it. Unlike other crawlers you may
    have used, this choice does _not_ affect the scope of the crawl.
- `Include`: An array of regular expressions that a URL must match in
    order to be crawler. If there is no valid `Include` expression,
    all discovered URLs could be within scope. Note that
    meta-characters must be double-escaped. Only meaningful in spider
    mode.
- `Exclude`: An array of regular expressions that filter the URLs to
    be crawled. Meta-characters must be double-escaped. Only meaningful
    in spider mode.
- `MaxDepth`: Only URLs fewer links than `MaxDepth` from the `From`
    list will be crawled.
- `WaitTime`: Pause time between spawning requests. Approximates crawl
    rate.  For instance, to crawl about 5 URLs per second, set this to
    "200ms". It uses Go's [time parsing
    rules](https://golang.org/pkg/time/#ParseDuration).
- `Connections`: The maximum number of current connections. If the
    configured value is < 1, it will be set to 1 upon starting the
    crawl.
- `UserAgent`: The user-agent to send with HTTP requests.
- `RobotsUserAgent`: The user-agent to test robots.txt rules against.
- `RespectNofollow`: If this is true, links with a `rel="nofollow"`
    attribute will not be included in the crawl.
- `Header`: An array of objects with properties "K" and "V",
    signifying key/value pairs to be added to all requests.
	
The `MaxDepth`, `Include`, and `Exclude` options only apply to spider
mode.
	
## Summarizing crawl scope

Given your specified Include and Exclude lists, defined above, here
is how the crawler decides whether a URL is in scope:

1. If the URL matches a rule in the Exclude list, it will not be crawled.
2. If the URL matches a rule in the Include list, it will be crawled.
3. If the URL matches neither the Exclude nor Include list, then if the
    Include list is empty, it will be crawled, but if the Include list
	is _not_ empty, it will not be crawled.

Note that only one of these cases will apply (as in Go's switch
statement, by way of analogy).

Finally, no URLs will be in scope if they are further than `MaxDepth`
links from the `From` set of URLs.

## Use with BigQuery

Run `crawl schema >schema.json` to get a BigQuery-compatible schema
definition file. The file is automatically generated (via `go
generate`) from the structure of the result object generated by the
crawler, so it should always be up-to-date.

If you find an incompatibility between the output schema file and the
data produced from a crawl, please flag as a bug on GitHub.

Crawl files can be large, and it is convenient to upload them directly
to Google Cloud Storage without storing them locally. This can be done
by piping the output of `crawl` to `gsutil`:

```sh
crawl spider config.json | gsutil cp - gs://my-bucket/crawl-data.txt
```

## Bugs, errors, contributions

All reports, requests, and contributions are welcome. Please handle
all of them through the GitHub repository. Thank you!

## License

MIT
