package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type Store struct{ db *sql.DB }

type scanner interface{ Scan(...any) error }

func scanProfile(row scanner) (Profile, error) {
	var p Profile
	err := row.Scan(&p.ID, &p.Name, &p.Title, &p.Description, &p.CoverPhotoURL, &p.ProfilePhotoURL, &p.Email, &p.Location, &p.GithubURL, &p.LinkedinURL)
	return p, err
}

const profileColumns = `id, name, title, description, cover_photo_url, profile_photo_url, email, location, github_url, linkedin_url`

func (s *Store) profile(ctx context.Context) (Profile, error) {
	return scanProfile(s.db.QueryRowContext(ctx, `SELECT `+profileColumns+` FROM profile ORDER BY id LIMIT 1`))
}

func (s *Store) saveProfile(ctx context.Context, p Profile) (Profile, error) {
	q := `UPDATE profile SET name=$1, title=$2, description=$3, cover_photo_url=$4, profile_photo_url=$5, email=$6, location=$7, github_url=$8, linkedin_url=$9 WHERE id=(SELECT id FROM profile ORDER BY id LIMIT 1) RETURNING ` + profileColumns
	args := []any{p.Name, p.Title, p.Description, p.CoverPhotoURL, p.ProfilePhotoURL, p.Email, p.Location, p.GithubURL, p.LinkedinURL}
	result, err := scanProfile(s.db.QueryRowContext(ctx, q, args...))
	if errors.Is(err, sql.ErrNoRows) {
		q = `INSERT INTO profile (name,title,description,cover_photo_url,profile_photo_url,email,location,github_url,linkedin_url) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING ` + profileColumns
		return scanProfile(s.db.QueryRowContext(ctx, q, args...))
	}
	return result, err
}

func (s *Store) projects(ctx context.Context) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.id,p.title,p.description,p.image_url,p.live_url,p.repo_url,p.display_order, ARRAY(SELECT t.technologies FROM project_technologies t WHERE t.project_id=p.id ORDER BY t.ctid) FROM project p ORDER BY p.display_order,p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.ImageURL, &p.LiveURL, &p.RepoURL, &p.DisplayOrder, pq.Array(&p.Technologies)); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (s *Store) education(ctx context.Context) ([]Education, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,institution,degree,field_of_study,start_year,end_year,description,display_order FROM education ORDER BY display_order,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Education{}
	for rows.Next() {
		var e Education
		if err := rows.Scan(&e.ID, &e.Institution, &e.Degree, &e.FieldOfStudy, &e.StartYear, &e.EndYear, &e.Description, &e.DisplayOrder); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *Store) experiences(ctx context.Context) ([]Experience, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT e.id,e.company,e.role,e.location,e.start_date,e.end_date,e.display_order, e.company_logo_path, ARRAY(SELECT h.highlights FROM experience_highlights h WHERE h.experience_id=e.id ORDER BY h.ctid) FROM experience e ORDER BY e.display_order,e.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Experience{}
	for rows.Next() {
		var e Experience
		if err := rows.Scan(&e.ID, &e.Company, &e.Role, &e.Location, &e.StartDate, &e.EndDate, &e.DisplayOrder, &e.CompanyLogoPath, pq.Array(&e.Highlights)); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *Store) portfolio(ctx context.Context) (Portfolio, error) {
	var p Portfolio
	var err error
	if p.Profile, err = s.profile(ctx); err != nil {
		return p, err
	}
	if p.Projects, err = s.projects(ctx); err != nil {
		return p, err
	}
	if p.Education, err = s.education(ctx); err != nil {
		return p, err
	}
	if p.Experience, err = s.experiences(ctx); err != nil {
		return p, err
	}
	err = s.db.QueryRowContext(ctx, "SELECT (SELECT image_path FROM education_photo WHERE id=1)").Scan(&p.EducationPhotoPath)
	return p, err
}

func (s *Store) saveProject(ctx context.Context, p Project, id int64) (Project, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return p, err
	}
	defer tx.Rollback()
	args := []any{p.Title, p.Description, p.ImageURL, p.LiveURL, p.RepoURL, p.DisplayOrder}
	var row *sql.Row
	if id == 0 {
		row = tx.QueryRowContext(ctx, `INSERT INTO project (title,description,image_url,live_url,repo_url,display_order) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, args...)
	} else {
		args = append(args, id)
		row = tx.QueryRowContext(ctx, `UPDATE project SET title=$1,description=$2,image_url=$3,live_url=$4,repo_url=$5,display_order=$6 WHERE id=$7 RETURNING id`, args...)
	}
	if err = row.Scan(&p.ID); err != nil {
		return p, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM project_technologies WHERE project_id=$1`, p.ID); err != nil {
		return p, err
	}
	for _, value := range p.Technologies {
		if _, err = tx.ExecContext(ctx, `INSERT INTO project_technologies (project_id,technologies) VALUES ($1,$2)`, p.ID, value); err != nil {
			return p, err
		}
	}
	if p.Technologies == nil {
		p.Technologies = []string{}
	}
	return p, tx.Commit()
}

func (s *Store) saveEducation(ctx context.Context, e Education, id int64) (Education, error) {
	args := []any{e.Institution, e.Degree, e.FieldOfStudy, e.StartYear, e.EndYear, e.Description, e.DisplayOrder}
	var row *sql.Row
	if id == 0 {
		row = s.db.QueryRowContext(ctx, `INSERT INTO education (institution,degree,field_of_study,start_year,end_year,description,display_order) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, args...)
	} else {
		args = append(args, id)
		row = s.db.QueryRowContext(ctx, `UPDATE education SET institution=$1,degree=$2,field_of_study=$3,start_year=$4,end_year=$5,description=$6,display_order=$7 WHERE id=$8 RETURNING id`, args...)
	}
	err := row.Scan(&e.ID)
	return e, err
}

func (s *Store) saveExperience(ctx context.Context, e Experience, id int64) (Experience, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return e, err
	}
	defer tx.Rollback()
	args := []any{e.Company, e.Role, e.Location, e.StartDate, e.EndDate, e.DisplayOrder}
	var row *sql.Row
	if id == 0 {
		row = tx.QueryRowContext(ctx, `INSERT INTO experience (company,role,location,start_date,end_date,display_order) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id,company_logo_path`, args...)
	} else {
		args = append(args, id)
		row = tx.QueryRowContext(ctx, `UPDATE experience SET company=$1,role=$2,location=$3,start_date=$4,end_date=$5,display_order=$6 WHERE id=$7 RETURNING id,company_logo_path`, args...)
	}
	if err = row.Scan(&e.ID, &e.CompanyLogoPath); err != nil {
		return e, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM experience_highlights WHERE experience_id=$1`, e.ID); err != nil {
		return e, err
	}
	for _, value := range e.Highlights {
		if _, err = tx.ExecContext(ctx, `INSERT INTO experience_highlights (experience_id,highlights) VALUES ($1,$2)`, e.ID, value); err != nil {
			return e, err
		}
	}
	if e.Highlights == nil {
		e.Highlights = []string{}
	}
	return e, tx.Commit()
}

func (s *Store) deleteItem(ctx context.Context, kind string, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var table string
	switch kind {
	case "projects":
		table = "project"
		_, err = tx.ExecContext(ctx, `DELETE FROM project_technologies WHERE project_id=$1`, id)
	case "education":
		table = "education"
	case "experience":
		table = "experience"
		_, err = tx.ExecContext(ctx, `DELETE FROM experience_highlights WHERE experience_id=$1`, id)
	default:
		return fmt.Errorf("unknown item type")
	}
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE id=$1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}
