package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const adminIterations = 600000

func (s *Store) seed(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(72155422)`); err != nil {
		return err
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM profile`).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return tx.Commit()
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO profile (name,title,description,cover_photo_url,profile_photo_url,email,location,github_url,linkedin_url)
VALUES ('Aditya','Software Engineer', 'I''m a developer who enjoys building clean, reliable software and turning ideas into polished products. This is placeholder text: replace it with a short paragraph about who you are, what you work on, and what you care about.', '/aditya_cover.jpg','/aditya_profile.jpg','adityagupta4114@gmail.com','India','https://github.com/disconnect101','https://www.linkedin.com/in/adityagupta4114/');
WITH p AS (INSERT INTO project (title,description,repo_url,display_order) VALUES ('Project One','Short description of what this project does and the problem it solves.','https://github.com/your-username/project-one',1) RETURNING id) INSERT INTO project_technologies SELECT id, unnest(ARRAY['Java','Spring Boot','PostgreSQL']) FROM p;
WITH p AS (INSERT INTO project (title,description,repo_url,live_url,display_order) VALUES ('Project Two','Short description of another project. Mention your role and the outcome.','https://github.com/your-username/project-two','https://example.com',2) RETURNING id) INSERT INTO project_technologies SELECT id, unnest(ARRAY['React','TypeScript','Vite']) FROM p;
WITH p AS (INSERT INTO project (title,description,display_order) VALUES ('Project Three','Short description of a third project, such as a 3D model or a side project.',3) RETURNING id) INSERT INTO project_technologies SELECT id, unnest(ARRAY['Blender','3D Printing']) FROM p;
INSERT INTO education (institution,degree,field_of_study,start_year,end_year,description,display_order) VALUES ('Your University','Bachelor of Technology','Computer Science','2018','2022','Relevant coursework, honors, or activities.',1), ('Your School','Higher Secondary','Science','2016','2018',NULL,2);
WITH e AS (INSERT INTO experience (company,role,location,start_date,end_date,display_order) VALUES ('Company Name','Software Engineer','Bengaluru, India','Jul 2022','Present',1) RETURNING id) INSERT INTO experience_highlights SELECT id, unnest(ARRAY['Built and shipped features used by thousands of users.','Improved API response times by refactoring slow queries.','Mentored junior developers and reviewed code.']) FROM e;
WITH e AS (INSERT INTO experience (company,role,location,start_date,end_date,display_order) VALUES ('Previous Company','Software Engineering Intern','Remote','Jan 2022','Jun 2022',2) RETURNING id) INSERT INTO experience_highlights SELECT id, unnest(ARRAY['Developed internal tooling in Java.','Wrote unit and integration tests.']) FROM e;
`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) bootstrapAdmin(ctx context.Context, key string) error {
	if len(key) < 16 {
		return nil
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM admin_credential WHERE id=1)`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	hash := pbkdf2.Key([]byte(key), salt, adminIterations, 32, sha256.New)
	_, err := s.db.ExecContext(ctx, `INSERT INTO admin_credential (id,salt,password_hash,iterations) VALUES (1,$1,$2,$3) ON CONFLICT (id) DO NOTHING`, base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(hash), adminIterations)
	return err
}

func (s *Store) authenticate(ctx context.Context, key string) (bool, error) {
	var salt64, hash64 string
	var iterations int
	err := s.db.QueryRowContext(ctx, `SELECT salt,password_hash,iterations FROM admin_credential WHERE id=1`).Scan(&salt64, &hash64, &iterations)
	if errors.Is(err, sql.ErrNoRows) {
		return false, sql.ErrNoRows
	}
	if err != nil {
		return false, err
	}
	salt, err := base64.StdEncoding.DecodeString(salt64)
	if err != nil {
		return false, fmt.Errorf("invalid stored admin salt")
	}
	hash, err := base64.StdEncoding.DecodeString(hash64)
	if err != nil || len(hash) != 32 || iterations < 1 || iterations > 5000000 {
		return false, fmt.Errorf("invalid stored admin hash")
	}
	actual := pbkdf2.Key([]byte(key), salt, iterations, len(hash), sha256.New)
	return subtle.ConstantTimeCompare(hash, actual) == 1, nil
}
