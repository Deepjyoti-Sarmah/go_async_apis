package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ReportStore struct {
	db *sqlx.DB
}

func NewReportStore(db *sql.DB) *ReportStore {
	return &ReportStore{
		sqlx.NewDb(db, "postgres"),
	}
}

type Report struct {
	UserId               uuid.UUID  `db:"user_id"`
	Id                   uuid.UUID  `db:"id"`
	ReportType           string     `db:"report_type"`
	OutputFilePath       string     `db:"output_file_path"`
	DownloadUrl          *string    `db:"download_url"`
	DownloadUrlExpiresAt *time.Time `db:"download_url_expires_at"`
	ErrorMessage         *string    `db:"error_message"`
	CreatedAt            time.Time  `db:"created_at"`
	StartedAt            *time.Time `db:"started_at"`
	CompletedAt          *time.Time `db:"completed_at"`
	FailedAt             *time.Time `db:"failed_at"`
}

func (s *ReportStore) Create(ctx context.Context, userId uuid.UUID, reportType string) (*Report, error) {
	const insert = `INSERT INTO reports (user_id, report_type) VALUES ($1, $2) RETURNING *;`
	var report Report

	if err := s.db.GetContext(ctx, &report, insert, userId, reportType); err != nil {
		return nil, fmt.Errorf("failed to insert report for user %s: %w", userId, err)
	}

	return &report, nil
}


