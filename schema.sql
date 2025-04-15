
-- Типы ENUM
CREATE TYPE public.lessons_status AS ENUM ('пропущен', 'проведен', 'запланирован', 'отменен');
CREATE TYPE public.payment_status AS ENUM ('оплачен', 'не оплачен', 'предоплата');

-- Таблицы
CREATE TABLE public.students (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    full_name text NOT NULL,
    phone_number text NOT NULL,
    parent_name text NOT NULL,
    parent_phone_number text NOT NULL,
    curator_id uuid,
    courses text[],
    platform_link text,
    crm_link text,
    created_at timestamp without time zone DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE public.lessons (
    id uuid NOT NULL,
    student_id uuid,
    date date NOT NULL,
    feedback text,
    payment_status public.payment_status,
    lessons_status public.lessons_status,
    feedback_date timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (student_id) REFERENCES public.students(id) ON DELETE CASCADE
);

CREATE TABLE public.roles (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    permissions jsonb NOT NULL
);

CREATE TABLE public.users (
    id serial PRIMARY KEY,
    email text NOT NULL UNIQUE,
    password text NOT NULL,
    role_id integer REFERENCES public.roles(id) ON DELETE SET NULL,
    created_at timestamp without time zone DEFAULT now(),
    reset_token text,
    reset_token_expires_at timestamp without time zone
);

CREATE TABLE public.sessions (
    id serial PRIMARY KEY,
    user_id integer REFERENCES public.users(id) ON DELETE CASCADE,
    refresh_token text NOT NULL UNIQUE,
    expires_at timestamp without time zone NOT NULL
);
