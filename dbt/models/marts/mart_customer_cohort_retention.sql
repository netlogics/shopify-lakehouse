-- Monthly cohort retention matrix: for each acquisition cohort (month of a
-- customer's first order), what fraction of that cohort placed at least
-- one order in each subsequent month (period_number = months since first
-- order; 0 = the acquisition month itself). Excludes synthetic-fraud-
-- flagged order lines (see shopify-lakehouse-92x.2).

with customer_orders as (
    select
        customer_id,
        order_id,
        cast(created_at as date) as order_date
    from {{ ref('stg_order_details') }}
    where not is_synthetic_fraud and customer_id is not null
),

first_order as (
    select
        customer_id,
        min(order_date) as first_order_date
    from customer_orders
    group by customer_id
),

cohort_activity as (
    select distinct
        f.customer_id,
        date_trunc('month', f.first_order_date) as cohort_month,
        date_trunc('month', co.order_date) as activity_month
    from customer_orders co
    join first_order f on f.customer_id = co.customer_id
),

period_activity as (
    select
        customer_id,
        cohort_month,
        (extract(year from activity_month) - extract(year from cohort_month)) * 12
            + (extract(month from activity_month) - extract(month from cohort_month)) as period_number
    from cohort_activity
),

cohort_sizes as (
    select cohort_month, count(distinct customer_id) as cohort_size
    from period_activity
    where period_number = 0
    group by cohort_month
)

select
    p.cohort_month,
    p.period_number,
    count(distinct p.customer_id) as active_customers,
    cs.cohort_size,
    round(100.0 * count(distinct p.customer_id) / nullif(cs.cohort_size, 0), 2) as retention_pct
from period_activity p
join cohort_sizes cs on cs.cohort_month = p.cohort_month
group by p.cohort_month, p.period_number, cs.cohort_size
