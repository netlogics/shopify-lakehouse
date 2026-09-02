-- Revenue rolled up by customer segment. The generator has no explicit
-- demographic/loyalty segment field on customers, so segments are derived
-- here from each customer's own net revenue via quartiles (NTILE), a
-- standard RFM-lite approach when no segment attribute exists upstream.
-- Excludes synthetic-fraud-flagged order lines (see shopify-lakehouse-92x.2).

with customer_revenue as (
    select
        customer_id,
        count(distinct order_id) as order_count,
        sum(price * quantity) as gross_revenue,
        sum(price * quantity) - sum(total_discount) as net_revenue
    from {{ ref('stg_order_details') }}
    where not is_synthetic_fraud and customer_id is not null
    group by customer_id
),

segmented as (
    select
        *,
        ntile(4) over (order by net_revenue desc) as revenue_quartile
    from customer_revenue
)

select
    case revenue_quartile
        when 1 then 'VIP'
        when 2 then 'High'
        when 3 then 'Medium'
        else 'Low'
    end as revenue_segment,
    count(*) as customer_count,
    sum(order_count) as total_orders,
    sum(gross_revenue) as segment_gross_revenue,
    sum(net_revenue) as segment_net_revenue,
    avg(net_revenue) as avg_net_revenue_per_customer
from segmented
group by revenue_quartile
