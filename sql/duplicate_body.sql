/* 
Group all pages by BodyTextHash, and report how many share that
value. The resulting rows each represent a group of pages with
duplicated body content.
*/

WITH q AS (SELECT * FROM crawl)

SELECT ARRAY_AGG(DISTINCT Address.Full) AS Examples,
       BodyTextHash,
       COUNT(*) AS N
FROM q
GROUP BY BodyTextHash
ORDER BY N DESC
