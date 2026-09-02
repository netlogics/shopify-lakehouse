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
    published_at
from {{ source('lakehouse', 'products') }}
