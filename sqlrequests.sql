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


create table lessons(
    id UUID PRIMARY KEY,
    student_id UUID REFERENCES students(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    feedback TEXT,
    payment_status payment_status,
    lessons_status lessons_status,
    feedback_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
)
CREATE TYPE payment_status AS ENUM ('оплачен', 'не оплачен', 'предоплата');

CREATE TYPE lessons_status AS ENUM ('пропущен', 'проведен', 'запланирован', 'отменен');

DROP TYPE IF EXISTS lessons_status


create table courses(
  id UUID PRIMARY KEY,
  text TEXT NOT null
)


_______________________________________________________________________
//Образец для Post lessons
{
  "student_id": "88a57423-8b7f-4977-9479-7aefecda6fce",
  "course_id": "c3ad5757-27be-4958-bc63-d9d20a4c53c6",
  "date": "01.04.2025",
  "feedback": "Отличный урок!",
  "feedback_date": "01.04.2025",
  "created_at": "01.04.2025"
}



//в put запросе я добавила
{
  "payment_status":"оплачен",
  "lessons_status":"проведен"
} 

_____________________________________________________________________
//POST lk
/* localhost:8081/students */
{
    "full_name": "Молдир Берикканова",
    "phone_number": "+77086108823",
    "parent_name": "Анна Берикканова",
    "parent_phone_number": "+77081234567",
    "courses": ["math", "physics", "chemistry"],
    "platform_link": "https://platform.example.com",
    "crm_link": "https://crm.example.com",
    "created_at": "27.03.2025"
  }

//доработка таблицы уроков
alter table lessons add column course_id UUID REFERENCES courses(id) ON DELETE CASCADE
