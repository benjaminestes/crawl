-- List all addresses and whether they have a noindex tag
SELECT Address,
       REGEXP_CONTAINS(Robots, r"\bnoindex\b") AS Noindex
FROM crawl
