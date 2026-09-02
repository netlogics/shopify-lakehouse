-- Current stock status per (inventory_item_id, location_id): the single
-- most recent available count, flagged out_of_stock / low_stock / healthy.
-- low_stock_threshold is a dbt var (dbt_project.yml), overridable per run
-- via --vars.

with latest as (
    select
        inventory_item_id,
        location_id,
        available,
        updated_at,
        row_number() over (
            partition by inventory_item_id, location_id
            order by updated_at desc
        ) as rn
    from {{ ref('stg_inventory_levels') }}
)

select
    l.inventory_item_id,
    l.location_id,
    l.available,
    l.updated_at as as_of,
    v.variant_id,
    v.product_id,
    p.title as product_title,
    v.sku,
    case
        when l.available <= 0 then 'out_of_stock'
        when l.available <= {{ var('low_stock_threshold') }} then 'low_stock'
        else 'healthy'
    end as stock_status
from latest l
left join {{ ref('stg_product_variants') }} v on v.inventory_item_id = l.inventory_item_id
left join {{ ref('stg_products') }} p on p.product_id = v.product_id
where l.rn = 1
