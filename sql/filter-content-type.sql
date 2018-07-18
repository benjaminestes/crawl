SELECT
    FullAddress,
    b.Val AS Type
FROM
    `{{ table }}` AS a,
    UNNEST(Response.Header) AS b
WHERE
    Response.StatusCode = 200
    AND b.Key = "Content-Type"
    -- Your Content-Type here
    AND b.Val LIKE "%html%"
