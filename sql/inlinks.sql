-- Produces a table of information about pages which are linkes to by
-- pages that are a part of the crawl.
WITH
	-- q is an alias for your crawl data
	q AS (SELECT * FROM crawl),

	-- r is information about addresses that appear in links.
	-- These addresses may or may not themselves have been
	-- included in the original crawl.
	r AS (
     	SELECT
		link.Address.Full AS FullAddress,
          	COUNT(q.Address) AS InLinks
	FROM q, UNNEST(Links) AS link
	GROUP BY link.Address )

-- Result is a table with information about addresses appearing as
-- links.  Because it's possible a page is linked to without being in
-- the set of crawled pages, some rows may only include an address.
SELECT
	FullAddress,
       	StatusCode,
       	InLinks
FROM r LEFT JOIN q ON r.FullAddress = q.Address.Full
ORDER BY InLinks DESC
