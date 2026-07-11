package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
)

type newsletterRepository struct {
	db *sqlx.DB
}

func NewNewsletterRepository(db *sqlx.DB) repository.NewsletterRepository {
	return &newsletterRepository{db: db}
}

func (r *newsletterRepository) FindAll(ctx context.Context, filter repository.SubscriberFilter) ([]entity.NewsletterSubscriber, int64, error) {
	var conditions []string
	var args []interface{}

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(email LIKE ? OR name LIKE ?)")
		q := "%" + filter.Search + "%"
		args = append(args, q, q)
	}

	where := "1=1"
	if len(conditions) > 0 {
		where = strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM newsletter_subscribers WHERE %s`, where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	listQuery := fmt.Sprintf(`
		SELECT * FROM newsletter_subscribers
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)
	listArgs := append(append([]interface{}{}, args...), filter.PerPage, offset)

	var subs []entity.NewsletterSubscriber
	if err := r.db.SelectContext(ctx, &subs, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

func (r *newsletterRepository) FindByEmail(ctx context.Context, email string) (*entity.NewsletterSubscriber, error) {
	var sub entity.NewsletterSubscriber
	if err := r.db.GetContext(ctx, &sub, `SELECT * FROM newsletter_subscribers WHERE email = ?`, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrSubscriberNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *newsletterRepository) FindByID(ctx context.Context, id string) (*entity.NewsletterSubscriber, error) {
	var sub entity.NewsletterSubscriber
	if err := r.db.GetContext(ctx, &sub, `SELECT * FROM newsletter_subscribers WHERE id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrSubscriberNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *newsletterRepository) Create(ctx context.Context, sub *entity.NewsletterSubscriber) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO newsletter_subscribers (id, email, name, source, status)
		VALUES (:id, :email, :name, :source, :status)
	`, sub)
	return err
}

func (r *newsletterRepository) UpdateStatus(ctx context.Context, id string, status entity.SubscriberStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE newsletter_subscribers SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *newsletterRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM newsletter_subscribers WHERE id = ?`, id)
	return err
}
