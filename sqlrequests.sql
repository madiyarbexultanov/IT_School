create table students(
    id UUID PRIMARY key,
    full_name TEXT NOT null,
    phone_number TEXT NOT null,
    parent_name TEXT NOT null,
    parent_phone_number TEXT NOT null,
    curator_id UUID REFERENCES curators(id),
    courses TEXT[],
    platform_link TEXT,
    crm_link TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);


create table lessonss(
    id UUID PRIMARY KEY,
    student_id UUID REFERENCES students(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    feedback TEXT,
    status payment_status AS ENUM ('оплачен', 'не оплачен', 'предоплата'),
    feedback_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
)



//Образец для Post
{
  "student_id": "550e8400-e29b-41d4-a716-446655440000",
  "date": "2025-04-01",
  "feedback": "Отличный урок!",
  "status": "проведен",
  "feedback_date": "01.04.2025",
  "created_at": "01.04.2025"
}
