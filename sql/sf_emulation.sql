-- Outputs a bunch of the information you'd get from running a
-- ScremingFrog crawl. It's not a 1:1 match, so you can't copy/paste
-- the output into a context expecting SF output and expect it to
-- work.
WITH
        q AS ( -- extend                                                                                                                                                                      
        SELECT
                *,
                COALESCE(Address.Full != Canonical.Address.Full, true) AS HasOtherCanonical,
                REGEXP_CONTAINS(Robots, r"\bnoindex\b") AS Noindex,
                REGEXP_CONTAINS(Robots, r"\bnofollow\b") AS Nofollow
        FROM crawl ), -- your crawl here                                                                                                                                                       

        r AS ( -- count links to each page                                                                                                                                                    
        SELECT DISTINCT
                target.Address.Full AS FullAddress,
                COUNT(DISTINCT source.Address.Full) OVER (PARTITION BY target.Address.Full) AS InLinks
        FROM q AS source, UNNEST(Links) AS target )

SELECT DISTINCT
        Depth,
        r.FullAddress,
        (SELECT V FROM UNNEST(Header) WHERE K = "Content-Type") AS ContentType,
        Status,
        StatusCode,
        Title,
        COUNT(*) OVER (PARTITION BY Title) AS TitleCount,
        length(Title) AS TitleLength,
        H1,
        length(H1) AS H1Length,
        Canonical.Address.Full AS Canonical,
        Description,
        Robots,
        Noindex,
        Nofollow,
        NOT (StatusCode != 200 OR Noindex OR HasOtherCanonical) AS Indexable,
        InLinks,
        BodyTextHash,
        COUNT(*) OVER (PARTITION BY BodyTextHash) AS BodyCount
FROM q LEFT JOIN r ON q.Address.Full = r.FullAddress
WHERE q.Address IS NOT NULL
ORDER BY Depth ASC
