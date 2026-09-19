ALTER TABLE plan_categories
    ADD CONSTRAINT plan_categories_daily_plan_category_unique UNIQUE (daily_plan_id, category_id);
