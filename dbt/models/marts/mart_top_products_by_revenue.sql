-- Product-level revenue ranking. Excludes synthetic-fraud-flagged order
-- lines (see shopify-lakehouse-92x.2) so a handful of fraud-injected bulk
-- orders don't distort which products are genuinely top sellers.
--
-- A meaningful share of order_details rows have product_id = null (no
-- product reference at all, distinct from a reference to a deleted
-- product) -- these are grouped into an explicit 'Unknown Product' row
-- rather than dropped, since it's real revenue that still needs surfacing.

select
    od.product_id,
    coalesce(p.title, 'Unknown Product') as title,
    p.vendor,
    p.product_type,
    sum(od.quantity) as units_sold,
    sum(od.price * od.quantity) as gross_revenue,
    sum(od.price * od.quantity) - sum(od.total_discount) as net_revenue,
    row_number() over (
        order by sum(od.price * od.quantity) - sum(od.total_discount) desc
    ) as revenue_rank
from {{ ref('stg_order_details') }} od
left join {{ ref('stg_products') }} p on p.product_id = od.product_id
where not od.is_synthetic_fraud
group by od.product_id, p.title, p.vendor, p.product_type
