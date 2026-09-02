-- Product catalog summary: one row per product with variant-level rollups.
-- First mart built by the dbt scaffold -- proves the Dremio SQL write path
-- (CTAS into nessie.marts) works end-to-end, not just staging passthroughs.
-- Sales/revenue, inventory health, and customer analytics marts are built
-- out separately (see shopify-lakehouse-1ih.2/.3/.4).

with variants as (
    select
        product_id,
        count(*) as variant_count,
        min(price) as min_price,
        max(price) as max_price,
        sum(inventory_quantity) as total_inventory_quantity
    from {{ ref('stg_product_variants') }}
    group by product_id
)

select
    p.product_id,
    p.title,
    p.vendor,
    p.product_type,
    p.status,
    p.created_at,
    p.updated_at,
    coalesce(v.variant_count, 0) as variant_count,
    v.min_price,
    v.max_price,
    coalesce(v.total_inventory_quantity, 0) as total_inventory_quantity
from {{ ref('stg_products') }} p
left join variants v on v.product_id = p.product_id
