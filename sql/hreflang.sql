/* 

Produces a table with one row per hreflang tag seen in the crawl.
This row includes the source address, the target address, whether the
hreflang relationship is reciprocated, the language of the hreflang
tag, and the status code of the target page.

*/

WITH
        q AS ( -- project
        SELECT
                Address.Full AS FullAddress,
                Hreflang,
                StatusCode
        FROM crawl ), -- your crawl here!
	
        /* 

	`r` represents all the URLs which are targets of some hreflang
        tag that was seen in a crawl. These URLs may or may not have
        been seen themselves.  FullAddress = address of a target page,
        SourceAddress = URL that targeted it, HreflangCode = language
        specified by tag that targeted it.

        */

	r AS ( -- reciprocal analysis
        SELECT DISTINCT
                source.FullAddress AS SourceAddress,
                target.Address.Full AS FullAddress,
                target.Hreflang AS HreflangCode
        FROM q AS source, UNNEST(Hreflang) AS target )

/* 

Create one row for each URL that is the target of an hreflang tag;
i.e., in this table information in `q` is information about the target
page, not the source page.  For the targeted URL, show: 

- the source address (which must exist), 
- the target, 
- the language of the hreflang tag, 
- whether the source address appears in the (possibly empty) 
  set of hreflang tags on the target page, and 
- the status code of the target address.

*/

SELECT DISTINCT
        SourceAddress,
        FullAddress AS TargetAddress,
        HreflangCode,
        SourceAddress IN
                (SELECT Address.Full FROM UNNEST(q.Hreflang))
                AS Reciprocated,
        q.StatusCode AS TargetStatusCode
FROM r LEFT JOIN q USING (FullAddress)

