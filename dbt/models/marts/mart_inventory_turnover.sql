-- Inventory turnover by product: units sold (all-time, excluding
-- synthetic-fraud-flagged order lines -- see shopify-lakehouse-92x.2) over
-- average on-hand stock across all of a product's variants and locations
-- (all-time average of the raw inventory_levels event log). A ratio well
-- above 1 means stock is moving fast relative to what's typically on hand;
-- near 0 means slow-moving/overstocked inventory.

with sales as (
    select
        product_id,
        sum(quantity) as units_sold
    from {{ ref('stg_order_details') }}
    where not is_synthetic_fraud and product_id is not null
    group by product_id
),

avg_stock as (
    select
        v.product_id,
        avg(il.available) as avg_available
    from {{ ref('stg_inventory_levels') }} il
    join {{ ref('stg_product_variants') }} v on v.inventory_item_id = il.inventory_item_id
    group by v.product_id
)

select
    coalesce(s.product_id, a.product_id) as product_id,
    p.title,
    coalesce(s.units_sold, 0) as units_sold,
    a.avg_available,
    case
        when a.avg_available is null or a.avg_available = 0 then null
        else coalesce(s.units_sold, 0) / a.avg_available
    end as turnover_ratio
from sales s
full outer join avg_stock a on a.product_id = s.product_id
left join {{ ref('stg_products') }} p on p.product_id = coalesce(s.product_id, a.product_id)
