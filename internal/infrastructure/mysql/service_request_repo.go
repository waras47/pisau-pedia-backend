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

type serviceRequestRepository struct {
	db *sqlx.DB
}

func NewServiceRequestRepository(db *sqlx.DB) repository.ServiceRequestRepository {
	return &serviceRequestRepository{db: db}
}

func (r *serviceRequestRepository) FindAll(ctx context.Context, filter repository.ServiceRequestFilter) ([]entity.ServiceRequest, int64, error) {
	var conditions []string
	var args []interface{}

	if filter.Type != "" {
		conditions = append(conditions, "type = ?")
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	where := "1=1"
	if len(conditions) > 0 {
		where = strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM service_requests WHERE %s`, where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PerPage
	listQuery := fmt.Sprintf(`
		SELECT * FROM service_requests
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)
	listArgs := append(append([]interface{}{}, args...), filter.PerPage, offset)

	var requests []entity.ServiceRequest
	if err := r.db.SelectContext(ctx, &requests, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *serviceRequestRepository) FindByID(ctx context.Context, id string) (*entity.ServiceRequest, error) {
	var req entity.ServiceRequest
	if err := r.db.GetContext(ctx, &req, `SELECT * FROM service_requests WHERE id = ?`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrServiceRequestNotFound
		}
		return nil, err
	}
	return &req, nil
}

func (r *serviceRequestRepository) Create(ctx context.Context, req *entity.ServiceRequest) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO service_requests (id, type, status, customer_name, customer_email, customer_phone, message, quoted_price, admin_notes)
		VALUES (:id, :type, :status, :customer_name, :customer_email, :customer_phone, :message, :quoted_price, :admin_notes)
	`, req)
	return err
}

func (r *serviceRequestRepository) Update(ctx context.Context, req *entity.ServiceRequest) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE service_requests SET
			status = :status,
			quoted_price = :quoted_price,
			admin_notes = :admin_notes
		WHERE id = :id
	`, req)
	return err
}
