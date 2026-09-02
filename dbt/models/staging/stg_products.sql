-- nessie.lakehouse.products is an append-only event log: Flink INSERT INTO
-- never updates or deletes, so every product update (e.g. a status or
-- price change) lands as a NEW row with the same id. Deduplicated here to
-- the latest event per product_id so this staging model represents
-- current state, not a full change history.

with ranked as (
    select
        event_id,
        id as product_id,
        title,
        vendor,
        product_type,
        handle,
        status,
        tags,
        created_at,
        updated_at,
        published_at,
        row_number() over (partition by id order by updated_at desc) as rn
    from {{ source('lakehouse', 'products') }}
)

select
    event_id,
    product_id,
    title,
    vendor,
    product_type,
    handle,
    status,
    tags,
    created_at,
    updated_at,
    published_at
from ranked
where rn = 1
