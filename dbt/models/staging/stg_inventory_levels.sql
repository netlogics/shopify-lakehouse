select
    event_id,
    inventory_item_id,
    location_id,
    available,
    updated_at
from {{ source('lakehouse', 'inventory_levels') }}
