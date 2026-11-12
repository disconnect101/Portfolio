-- Names and columns match Spring Boot/Hibernate's existing PostgreSQL schema.
CREATE TABLE IF NOT EXISTS education_photo (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    image_path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS profile
(
    id
    BIGINT
    GENERATED
    BY
    DEFAULT AS
    IDENTITY
    PRIMARY
    KEY,
    name
    VARCHAR
(
    255
), title VARCHAR
(
    255
), description VARCHAR
(
    2000
),
    cover_photo_url VARCHAR
(
    255
), profile_photo_url VARCHAR
(
    255
), email VARCHAR
(
    255
),
    location VARCHAR
(
    255
), github_url VARCHAR
(
    255
), linkedin_url VARCHAR
(
    255
)
    );
ALTER TABLE profile ADD COLUMN IF NOT EXISTS resume_path TEXT;

CREATE TABLE IF NOT EXISTS project
(
    id
    BIGINT
    GENERATED
    BY
    DEFAULT AS
    IDENTITY
    PRIMARY
    KEY,
    title
    VARCHAR
(
    255
), description VARCHAR
(
    1500
), image_url VARCHAR
(
    255
),
    live_url VARCHAR
(
    255
), repo_url VARCHAR
(
    255
), display_order INTEGER NOT NULL
    );
CREATE TABLE IF NOT EXISTS project_technologies
(
    project_id
    BIGINT
    NOT
    NULL
    REFERENCES
    project
(
    id
), technologies VARCHAR
(
    255
)
    );
CREATE TABLE IF NOT EXISTS education
(
    id
    BIGINT
    GENERATED
    BY
    DEFAULT AS
    IDENTITY
    PRIMARY
    KEY,
    institution
    VARCHAR
(
    255
), degree VARCHAR
(
    255
), field_of_study VARCHAR
(
    255
),
    start_year VARCHAR
(
    255
), end_year VARCHAR
(
    255
), description VARCHAR
(
    1000
), display_order INTEGER NOT NULL
    );
CREATE TABLE IF NOT EXISTS experience
(
    id
    BIGINT
    GENERATED
    BY
    DEFAULT AS
    IDENTITY
    PRIMARY
    KEY,
    company
    VARCHAR
(
    255
), role VARCHAR
(
    255
), location VARCHAR
(
    255
),
    start_date VARCHAR
(
    255
), end_date VARCHAR
(
    255
), display_order INTEGER NOT NULL
    );
-- Copy the old shared logo to each existing entry once when upgrading.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_attribute
                   WHERE attrelid = 'experience'::regclass AND attname = 'company_logo_path' AND NOT attisdropped) THEN
        ALTER TABLE experience ADD COLUMN company_logo_path TEXT;
        IF to_regclass('company_logo') IS NOT NULL THEN
            UPDATE experience e SET company_logo_path = l.image_path
            FROM company_logo l WHERE l.company_key = lower(trim(e.company));
        END IF;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS experience_highlights
(
    experience_id
    BIGINT
    NOT
    NULL
    REFERENCES
    experience
(
    id
), highlights VARCHAR
(
    500
)
    );
CREATE TABLE IF NOT EXISTS admin_credential
(
    id
    BIGINT
    PRIMARY
    KEY,
    salt
    VARCHAR
(
    255
) NOT NULL,
    password_hash VARCHAR
(
    255
) NOT NULL, iterations INTEGER NOT NULL
    );
