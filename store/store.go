package store

import "database/sql"

type Store struct {
	Users             *UserStore
	RefreshTokenStore *RefreshTokenStore
	ReportStore       *ReportStore
}

func New(db *sql.DB) *Store {
	return &Store{
		NewUserStore(db),
		NewRefreshTokenStore(db),
		NewReportStore(db),
	}
}
