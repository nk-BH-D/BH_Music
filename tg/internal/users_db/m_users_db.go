package users_db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	pq "github.com/lib/pq"
)

type PostgresUs struct {
	DB_us *sql.DB
}

func NewUsPostgres(db_us_url string, smoc, smic int) (*PostgresUs, error) {
	db, err := sql.Open("pgx", db_us_url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(smoc)
	db.SetMaxIdleConns(smic)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	p := &PostgresUs{DB_us: db}
	if err := p.ensureUsSchema(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PostgresUs) CloseUs() error {
	return p.DB_us.Close()
}

func (p *PostgresUs) ensureUsSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schrema := `
	CREATE TABLE IF NOT EXISTS users (
		user_id BIGINT PRIMARY KEY,
		chat_id BIGINT,
		full_name TEXT NOT NULL,
		login TEXT UNIQUE,
		status TEXT NOT NULL,
		file_id_list TEXT[] NOT NULL DEFAULT'{}', 
		created_ad TIMESTAMP WITH TIME ZONE DEFAULT now()
	);`
	_, err := p.DB_us.ExecContext(ctx, schrema)
	return err
}

func (p *PostgresUs) InsertUser(ctx context.Context, user_id, chat_id int64, status, login, name string) error {
	_, err := p.DB_us.ExecContext(
		ctx,
		"INSERT INTO users (user_id, chat_id, full_name, login, status) VALUES ($1,$2,$3,$4,$5)",
		user_id,
		chat_id,
		name,
		login,
		status,
	)
	return err
}

func (p *PostgresUs) GetUserStatus(ctx context.Context, user_id int64) (string, string, string, error) {
	var (
		login  string
		status string
		name   string
	)
	row := p.DB_us.QueryRowContext(
		ctx,
		"SELECT login, status, full_name FROM users WHERE user_id=$1",
		user_id,
	)
	if err := row.Scan(&login, &status, &name); err != nil {
		return "", "", "", err
	}
	return login, status, name, nil
}

func (p *PostgresUs) UpdateUserData(ctx context.Context, user_id int64, login, name string) error {
	_, err := p.DB_us.ExecContext(
		ctx,
		"UPDATE users SET login=$1 full_name=$2 WHERE user_id=$3",
		login,
		name,
		user_id,
	)
	return err
}

func (p *PostgresUs) UpdateFileIDList(ctx context.Context, user_id int64, file_id string) error {
	_, err := p.DB_us.ExecContext(
		ctx,
		"UPDATE users SET file_id_list = array_append(file_id_list, $1) WHERE user_id = $2",
		file_id, user_id,
	)
	return err
}

func (p *PostgresUs) GetFileIDList(ctx context.Context, user_id int64) ([]int64, error) {
	var fileIDL []int64
	row := p.DB_us.QueryRowContext(
		ctx,
		"SELECT file_id_list FROM users WHERE user_id = $1",
		user_id,
	)
	if err := row.Scan(pq.Array(&fileIDL)); err != nil {
		return nil, err
	}
	return fileIDL, nil
}
