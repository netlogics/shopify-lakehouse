-- Daily stock-level snapshot per (inventory_item_id, location_id): the
-- last available count recorded each day. inventory_levels is a raw event
-- log (one row per update, ~10/s), so this reduces it to one row per
-- item/location/day for trend reporting rather than exposing the full
-- event stream as a "mart".

with ranked as (
    select
        inventory_item_id,
        location_id,
        available,
        updated_at,
        cast(updated_at as date) as stock_date,
        row_number() over (
            partition by inventory_item_id, location_id, cast(updated_at as date)
            order by updated_at desc
        ) as rn
    from {{ ref('stg_inventory_levels') }}
)

select
    r.stock_date,
    r.inventory_item_id,
    r.location_id,
    r.available,
    r.updated_at as as_of,
    v.variant_id,
    v.product_id,
    p.title as product_title,
    v.sku
from ranked r
left join {{ ref('stg_product_variants') }} v on v.inventory_item_id = r.inventory_item_id
left join {{ ref('stg_products') }} p on p.product_id = v.product_id
where r.rn = 1
