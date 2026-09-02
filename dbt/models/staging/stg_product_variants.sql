-- Same append-only-event-log dedup as stg_products (see that model's
-- comment) -- nessie.lakehouse.product_variants gets a new row per
-- variant update, deduplicated here to the latest event per variant_id.

with ranked as (
    select
        event_id,
        product_id,
        variant_id,
        title,
        price,
        sku,
        "position",
        inventory_policy,
        compare_at_price,
        fulfillment_service,
        inventory_management,
        option1,
        option2,
        option3,
        taxable,
        barcode,
        weight,
        weight_unit,
        inventory_item_id,
        inventory_quantity,
        requires_shipping,
        created_at,
        updated_at,
        row_number() over (partition by variant_id order by updated_at desc) as rn
    from {{ source('lakehouse', 'product_variants') }}
)

select
    event_id,
    product_id,
    variant_id,
    title,
    price,
    sku,
    "position",
    inventory_policy,
    compare_at_price,
    fulfillment_service,
    inventory_management,
    option1,
    option2,
    option3,
    taxable,
    barcode,
    weight,
    weight_unit,
    inventory_item_id,
    inventory_quantity,
    requires_shipping,
    created_at,
    updated_at
from ranked
where rn = 1
