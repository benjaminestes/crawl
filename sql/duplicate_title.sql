-- Count the unique instances of each title from the crawl. Then,
-- produce a table with one row for each page, the title of the page,
-- and the number of pages that have that title in total.
WITH
	q AS (SELECT * FROM crawl),
	
	r AS (
	SELECT
		Title,
       	  	COUNT(*) AS N
	FROM q
   	GROUP BY Title )

SELECT
	Address.Full AS FullAddress,
	r.Title,
	r.N
FROM q
JOIN r
USING (Title)
WHERE
	r.Title != ""
	AND r.N > 1
	AND q.StatusCode = 200
ORDER BY
        r.N DESC,
	r.Title DESC
