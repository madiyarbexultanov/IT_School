-- Создание пользовательских типов
CREATE TYPE public."attendance_type" AS ENUM ('урок', 'заморозка', 'пролонгация');
CREATE TYPE public."is_active" AS ENUM ('активен', 'неактивен');
CREATE TYPE public."lessons_status" AS ENUM ('пропущен', 'проведен', 'запланирован', 'отменен');
CREATE TYPE public."payment_type" AS ENUM ('оплата', 'предоплата', 'доплата');

-- Создание таблиц
CREATE TABLE courses (
    id uuid NOT NULL PRIMARY KEY,
    title text NOT NULL
);

CREATE TABLE roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    name text NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb NULL
);

CREATE TABLE users (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    email text NOT NULL UNIQUE,
    password text NOT NULL,
    role_id uuid NULL REFERENCES roles(id) ON DELETE SET NULL,
    created_at timestamp DEFAULT now() NULL,
    reset_token text NULL,
    reset_token_expires_at timestamp NULL,
    full_name text NULL,
    phone_number text NULL
);

CREATE TABLE curators (
    user_id uuid NOT NULL PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    student_ids _uuid DEFAULT '{}'::uuid[] NULL,
    course_ids _uuid DEFAULT '{}'::uuid[] NULL
);

CREATE TABLE sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token text NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE students (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    full_name text NOT NULL,
    phone_number text NOT NULL,
    parent_name text NOT NULL,
    parent_phone_number text NOT NULL,
    curator_id uuid NULL REFERENCES curators(user_id) ON DELETE SET NULL,
    courses _text NULL,
    platform_link text NULL,
    crm_link text NULL,
    created_at timestamp DEFAULT now() NULL,
    is_active public."is_active" NULL,
    course_id uuid NULL REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE attendance (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    student_id uuid NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    type public."attendance_type" NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL
);

CREATE TABLE attendance_freezes (
    attendance_id uuid NOT NULL PRIMARY KEY REFERENCES attendance(id) ON DELETE CASCADE,
    start_date date NOT NULL,
    end_date date NOT NULL,
    comment text NULL
);

CREATE TABLE attendance_lessons (
    attendance_id uuid NOT NULL PRIMARY KEY REFERENCES attendance(id) ON DELETE CASCADE,
    curator_id uuid NOT NULL REFERENCES curators(user_id) ON DELETE CASCADE,
    date date NOT NULL,
    format text NULL,
    feedback text NULL,
    feedbackdate timestamptz DEFAULT now() NULL,
    lessons_status public."lessons_status" DEFAULT 'запланирован'::lessons_status NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL
);

CREATE TABLE attendance_prolongations (
    attendance_id uuid NOT NULL PRIMARY KEY REFERENCES attendance(id) ON DELETE CASCADE,
    payment_type public."payment_type" NOT NULL,
    date date NOT NULL,
    amount numeric NOT NULL,
    comment text NULL
);

CREATE TABLE lessons (
    id uuid NOT NULL PRIMARY KEY,
    student_id uuid NULL REFERENCES students(id) ON DELETE CASCADE,
    date date NOT NULL,
    feedback text NULL,
    payment_status public."payment_status" NULL,
    lessons_status public."lessons_status" NULL,
    feedback_date timestamp NOT NULL,
    created_at timestamp DEFAULT now() NULL,
    course_id uuid NULL REFERENCES courses(id) ON DELETE CASCADE
);