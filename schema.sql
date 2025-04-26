-- Create enum types
CREATE TYPE "is_active" AS ENUM ('активен', 'неактивен');
CREATE TYPE "lessons_status" AS ENUM ('пропущен', 'проведен', 'запланирован', 'отменен');
CREATE TYPE "payment_status" AS ENUM ('оплачен', 'не оплачен', 'предоплата');

-- Create tables
CREATE TABLE courses (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    title text NOT NULL
);

CREATE TABLE roles (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    name text NOT NULL,
    permissions jsonb DEFAULT '{}'::jsonb
);

CREATE TABLE users (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    email text NOT NULL UNIQUE,
    password text NOT NULL,
    role_id uuid REFERENCES roles(id) ON DELETE SET NULL,
    created_at timestamp DEFAULT now(),
    reset_token text,
    reset_token_expires_at timestamp,
    full_name text,
    phone_number text
);

CREATE TABLE curators (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    student_ids uuid[] DEFAULT '{}'::uuid[],
    course_ids uuid[] DEFAULT '{}'::uuid[]
);

CREATE TABLE students (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    full_name text NOT NULL,
    phone_number text NOT NULL,
    parent_name text NOT NULL,
    parent_phone_number text NOT NULL,
    curator_id uuid REFERENCES curators(user_id) ON DELETE SET NULL,
    courses text[],
    platform_link text,
    crm_link text,
    created_at timestamp DEFAULT now(),
    is_active "is_active",
    course_id uuid REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE lessons (
    id uuid PRIMARY KEY,
    student_id uuid REFERENCES students(id) ON DELETE CASCADE,
    date date NOT NULL,
    feedback text,
    payment_status "payment_status",
    lessons_status "lessons_status",
    feedback_date timestamp NOT NULL,
    created_at timestamp DEFAULT now(),
    course_id uuid REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE sessions (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token text NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP
);