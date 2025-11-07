create table public.pets_types (
  id uuid not null default gen_random_uuid (),
  name text not null,
  constraint pets_types_pkey primary key (id)
) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_pets_types_name on public.pets_types using btree (name) TABLESPACE pg_default;

create table public.pets_age_ranges (
  id uuid not null default gen_random_uuid (),
  name text not null,
  pet_type_id uuid not null,
  constraint pets_age_ranges_pet_type_id_fkey foreign KEY (pet_type_id) references pets_types (id) on delete RESTRICT,
  constraint pets_age_ranges_pkey primary key (id)
) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_pets_age_ranges_pet_type_id on public.pets_age_ranges using btree (pet_type_id) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_pets_age_ranges_name on public.pets_age_ranges using btree (name) TABLESPACE pg_default;

create table public.pets_sizes (
  id uuid not null default gen_random_uuid (),
  name text not null,
  pet_type_id uuid not null,
  constraint pets_sizes_pet_type_id_fkey foreign KEY (pet_type_id) references pets_types (id) on delete RESTRICT,
  constraint pets_sizes_pkey primary key (id)
) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_pets_sizes_pet_type_id on public.pets_sizes using btree (pet_type_id) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_pets_sizes_name on public.pets_sizes using btree (name) TABLESPACE pg_default;

create table public.pets_conditions (
  id uuid not null default gen_random_uuid (),
  name text not null,
  pet_type_id uuid not null,
  constraint pets_conditions_pet_type_id_fkey foreign KEY (pet_type_id) references pets_types (id) on delete RESTRICT,
  constraint pets_conditions_pkey primary key (id)
) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_pets_conditions_pet_type_id on public.pets_conditions using btree (pet_type_id) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_pets_conditions_name on public.pets_conditions using btree (name) TABLESPACE pg_default;

INSERT INTO public.pets_types (id, name) VALUES
  (gen_random_uuid(), LOWER('canine')),
  (gen_random_uuid(), LOWER('feline'));

INSERT INTO public.pets_age_ranges (id, name, pet_type_id) VALUES
  (gen_random_uuid(), LOWER('Cachorro (5m-1año)'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),
  (gen_random_uuid(), LOWER('Joven (1-6años)'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),
  (gen_random_uuid(), LOWER('Adulto (6-10años)'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),
  (gen_random_uuid(), LOWER('Senior (10-15+años)'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),

  (gen_random_uuid(), LOWER('Cachorro (5m-1año)'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
  (gen_random_uuid(), LOWER('Joven (1-5años)'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
  (gen_random_uuid(), LOWER('Adulto (5-7años)'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
  (gen_random_uuid(), LOWER('Senior (7-15+años)'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine')))
  ;

INSERT INTO public.pets_sizes (id, name, pet_type_id) VALUES
  (gen_random_uuid(), LOWER('Normal'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),
    (gen_random_uuid(), LOWER('Pequeñas'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
    (gen_random_uuid(), LOWER('Medianas'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
    (gen_random_uuid(), LOWER('Grandes'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine')));

INSERT INTO public.pets_conditions (id, name, pet_type_id) VALUES
  (gen_random_uuid(), LOWER('Macho Castrado'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
    (gen_random_uuid(), LOWER('Hembra Esterilizada'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
    (gen_random_uuid(), LOWER('Macho NO Castrado'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),
    (gen_random_uuid(), LOWER('Hembra NO Esterilizada'), (SELECT id FROM public.pets_types WHERE name = LOWER('canine'))),

    (gen_random_uuid(), LOWER('Macho Castrado'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),
    (gen_random_uuid(), LOWER('Hembra Esterilizada'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),
    (gen_random_uuid(), LOWER('Macho NO Castrado'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline'))),
    (gen_random_uuid(), LOWER('Hembra NO Esterilizada'), (SELECT id FROM public.pets_types WHERE name = LOWER('feline')));

create table public.pets (
  id uuid not null default gen_random_uuid (),
  user_id uuid not null,
  name text not null,
  breed text not null,
  gender text not null,
  weight numeric(5, 2) null,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  microchip_id text null,
  age_range_id uuid not null,
  condition_id uuid not null,
  size_id uuid not null,
    pet_type_id uuid not null,
  constraint pets_pkey primary key (id),
  constraint pets_user_id_fkey foreign KEY (user_id) references users (id) on delete CASCADE,
    constraint pets_age_range_id_fkey foreign KEY (age_range_id) references pets_age_ranges (id) on delete RESTRICT,
    constraint pets_condition_id_fkey foreign KEY (condition_id) references pets_conditions (id) on delete RESTRICT,
    constraint pets_size_id_fkey foreign KEY (size_id) references pets_sizes (id) on delete RESTRICT,
    constraint pets_pet_type_id_fkey foreign KEY (pet_type_id) references pets_types (id) on delete RESTRICT,
  constraint pets_gender_check check (
    (
      gender = any (array['male'::text, 'female'::text])
    )
  )
) TABLESPACE pg_default;

create index IF not exists idx_pets_user_id on public.pets using btree (user_id) TABLESPACE pg_default;

create trigger update_pets_updated_at BEFORE
update on pets for EACH row
execute FUNCTION update_updated_at ();

create table public.plans (
  id uuid not null default gen_random_uuid (),
  name text not null,
  monthly_price numeric(10, 2) not null,
  annual_limit numeric(10, 2) not null,
  description text null,
  shopify_id text null,
  pet_type_id uuid not null,
  created_at timestamp with time zone not null default now(),
  constraint plans_pkey primary key (id)
) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_plans_name on public.plans using btree (name) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_plans_shopify_id on public.plans using btree (shopify_id) TABLESPACE pg_default;
CREATE INDEX IF not exists idx_plans_pet_type_id on public.plans using btree (pet_type_id) TABLESPACE pg_default;

INSERT INTO public.plans (id, name, monthly_price, annual_limit, shopify_id, pet_type_id) VALUES
(gen_random_uuid(), LOWER('Plan 5000 feline + Bienestar'), 50.00, 5000.00, '8969215312122', (SELECT id FROM public.pets_types WHERE name = 'feline' LIMIT 1)),
    (gen_random_uuid(), LOWER('Plan 5000 feline'), 90.00, 10000.00, '8969215213818', (SELECT id FROM public.pets_types WHERE name = 'feline' LIMIT 1)),
    (gen_random_uuid(), LOWER('Plan 5000 canine + Bienestar'), 80.00, 10000.00, '8969215115514', (SELECT id FROM public.pets_types WHERE name = 'canine' LIMIT 1)),
    (gen_random_uuid(), LOWER('Plan 5000 canine'), 150.00, 10000.00, '8969215017210', (SELECT id FROM public.pets_types WHERE name = 'canine' LIMIT 1)),

  (gen_random_uuid(), LOWER('Plan 500 feline + Bienestar'), 50.00, 5000.00, '8969214951674', (SELECT id FROM public.pets_types WHERE name = 'feline' LIMIT 1)),
    (gen_random_uuid(), LOWER('Plan 500 feline'), 90.00, 10000.00, '8969214820602', (SELECT id FROM public.pets_types WHERE name = 'feline' LIMIT 1)),
    (gen_random_uuid(), LOWER('Plan 500 canine + Bienestar'), 80.00, 10000.00, '8969214722298', (SELECT id FROM public.pets_types WHERE name = 'canine' LIMIT 1)),
    (gen_random_uuid(), LOWER('Plan 500 canine'), 150.00, 10000.00, '8969214623994', (SELECT id FROM public.pets_types WHERE name = 'canine' LIMIT 1)),

    (gen_random_uuid(), LOWER('Plan 1500 feline + Bienestar'), 50.00, 5000.00, '8969214492922', (SELECT id FROM public.pets_types WHERE name = 'feline' LIMIT 1)),
    (gen_random_uuid(), LOWER('Plan 1500 feline'), 90.00, 10000.00, '8969214394618', (SELECT id FROM public.pets_types WHERE name = 'feline' LIMIT 1)),
        (gen_random_uuid(), LOWER('Plan 1500 canine + Bienestar'), 120.00, 20000.00, '8969214296314', (SELECT id FROM public.pets_types WHERE name = 'canine' LIMIT 1)),
        (gen_random_uuid(), LOWER('Plan 1500 canine'), 220.00, 20000.00, '8969214132474', (SELECT id FROM public.pets_types WHERE name = 'canine' LIMIT 1));

create table public.users (
  id uuid not null,
  name text not null,
  email text not null,
  phone text null,
  city text null,
  role public.app_role not null default 'user'::app_role,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  shopify_id text not null,
  constraint users_pkey primary key (id),
  constraint users_email_key unique (email),
  constraint users_shopify_id_key unique (shopify_id)
) TABLESPACE pg_default;

create index IF not exists idx_users_email on public.users using btree (email) TABLESPACE pg_default;

create index IF not exists idx_users_role on public.users using btree (role) TABLESPACE pg_default;

create trigger update_users_updated_at BEFORE
update on users for EACH row
execute FUNCTION update_updated_at ();

create table public.policies (
  id uuid not null default gen_random_uuid (),
  user_id uuid not null,
  pet_id uuid not null,
  plan_id uuid not null,
  start_date date not null,
  next_payment date not null,
  remaining_balance numeric(10, 2) not null,
  status text not null default 'active'::text,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  health_declared boolean not null default false,
  limit_period_start date not null default CURRENT_DATE,
  limit_period_end date not null default (CURRENT_DATE + '1 year'::interval),
  documents_verified boolean not null default false,
  constraint policies_pkey primary key (id),
  constraint policies_pet_id_fkey foreign KEY (pet_id) references pets (id) on delete CASCADE,
  constraint policies_plan_id_fkey foreign KEY (plan_id) references plans (id) on delete RESTRICT,
  constraint policies_user_id_fkey foreign KEY (user_id) references users (id) on delete CASCADE,
  constraint policies_status_check check (
    (
      status = any (
        array[
          'active'::text,
          'payment_pending'::text,
          'cancelled'::text,
          'pending_cancellation'::text
        ]
      )
    )
  )
) TABLESPACE pg_default;

create index IF not exists idx_policies_user_id on public.policies using btree (user_id) TABLESPACE pg_default;

create index IF not exists idx_policies_status on public.policies using btree (status) TABLESPACE pg_default;

create index IF not exists idx_policies_plan_id on public.policies using btree (plan_id) TABLESPACE pg_default;

create trigger trigger_auto_payment_status BEFORE
update on policies for EACH row
execute FUNCTION auto_update_payment_status ();

create trigger update_policies_updated_at BEFORE
update on policies for EACH row
execute FUNCTION update_updated_at ();

create table public.policies_payments (
  id uuid not null default gen_random_uuid (),
  policy_id uuid not null,
  payment_installment_id uuid not null,
  created_at timestamp with time zone not null default now(),
  constraint policies_payments_pkey primary key (id),
  constraint policies_payments_policy_id_payment_installment_id_key unique (policy_id, payment_installment_id),
  constraint policies_payments_payment_installment_id_fkey foreign KEY (payment_installment_id) references payment_installments (id) on delete CASCADE,
  constraint policies_payments_policy_id_fkey foreign KEY (policy_id) references policies (id) on delete CASCADE
) TABLESPACE pg_default;

create index IF not exists idx_policies_payments_policy_id on public.policies_payments using btree (policy_id) TABLESPACE pg_default;

create index IF not exists idx_policies_payments_installment_id on public.policies_payments using btree (payment_installment_id) TABLESPACE pg_default;

create table public.payment_installments (
  id uuid not null default gen_random_uuid (),
  installment_number integer not null,
  due_date date not null,
  amount numeric not null,
  status text not null default 'pending'::text,
  shopify_order_id text null,
  shopify_checkout_url text null,
  paid_at timestamp with time zone null,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  constraint payment_installments_pkey primary key (id),
  constraint payment_installments_status_check check (
    (
      status = any (
        array[
          'pending'::text,
          'paid'::text,
          'overdue'::text,
          'cancelled'::text
        ]
      )
    )
  )
) TABLESPACE pg_default;

create index IF not exists idx_installments_status on public.payment_installments using btree (status) TABLESPACE pg_default;

create index IF not exists idx_installments_due_date on public.payment_installments using btree (due_date) TABLESPACE pg_default;

create trigger update_payment_installments_updated_at BEFORE
update on payment_installments for EACH row
execute FUNCTION update_updated_at ();

create trigger update_policy_status_on_installment_change
after INSERT
or
update OF status on payment_installments for EACH row
execute FUNCTION update_policy_status_from_installment ();


DECLARE
  latest_installment RECORD;
BEGIN
  -- Iterate over all latest installments for the given policy
  FOR latest_installment IN
    SELECT
      policies_payments.policy_id,
      payment_installments.id AS installment_id,
      payment_installments.installment_number,
      payment_installments.due_date,
      payment_installments.status
    FROM policies_payments
    JOIN payment_installments ON policies_payments.payment_installment_id = payment_installments.id
    WHERE payment_installments.id = NEW.id
    ORDER BY payment_installments.installment_number DESC
  LOOP
    -- Update policy status based on installment status
    IF latest_installment.status = 'paid' THEN
      UPDATE policies
      SET status = 'active',
          next_payment = (latest_installment.due_date + INTERVAL '1 month')::DATE
      WHERE id = latest_installment.policy_id;
    ELSIF latest_installment.status = 'pending' OR latest_installment.status = 'overdue' THEN
      UPDATE policies
      SET status = 'payment_pending',
          next_payment = latest_installment.due_date
      WHERE id = latest_installment.policy_id;
    END IF;
  END LOOP;

  RETURN NEW;
END;
