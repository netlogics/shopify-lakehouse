select
    event_id,
    id as customer_id,
    email,
    first_name,
    last_name,
    phone,
    state,
    verified_email,
    tags,
    created_at,
    updated_at
from {{ source('lakehouse', 'customers') }}
