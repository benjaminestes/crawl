WITH
    q AS (
        SELECT
            a.FullAddress AS SourceAddress,
            b.FullAddress AS TargetAddress,
            b.Lang
        FROM
            `{{ table }}` AS a,
            UNNEST(Hreflang) AS b )
SELECT
    q.SourceAddress,
    q.TargetAddress,
    q.Lang,
    q.SourceAddress IN (
        SELECT
            FullAddress
        FROM
            UNNEST(r.Hreflang) ) AS Reciprocated
FROM
    q
    LEFT JOIN `{{ table }}` AS r
    ON
        q.TargetAddress = r.FullAddress
