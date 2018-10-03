-- Show whether an address has a self-referencing canonical tag.
SELECT Address,
       COALESCE(Address = Canonical.Address, false) AS SelfCanonical
FROM crawl
