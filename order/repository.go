package order

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, o Order) error
	GetOrderForAccount(ctx context.Context, accountID string) ([]*Order, error)
	ListOrders(ctx context.Context, skip uint64, take uint64) ([]Order, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) PutOrder(ctx context.Context, o Order) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	_, err = tx.ExecContext(ctx, "INSERT INTO orders(id, account_id, created_at, total_price) VALUES($1, $2, $3, $4)", o.ID, o.AccountID, o.CreatedAt, o.TotalPrice)
	if err != nil {
		return err
	}

	stmt, _ := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	defer stmt.Close()

	for _, p := range o.Products {
		_, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return err
		}
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil

}

func (r *postgresRepository) GetOrderForAccount(ctx context.Context, accountID string) ([]*Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
		o.id, 
		o.account_id, 
		o.created_at, 
		o.total_price::money::numeric::float8
		op.product_id, 
		op.quantity
		FROM orders o
		INNER JOIN order_products op ON o.id = op.order_id
		WHERE o.account_id = $1
		ORDER BY o.id DESC
		`,
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*Order)

	for rows.Next() {
		var (
			id, accountID string
			createdAt     time.Time
			totalPrice    float64
			productID     string
			quantity      uint64
		)
		if err = rows.Scan(&id, &accountID, &createdAt, &totalPrice, &productID, &quantity); err != nil {
			return nil, err
		}

		order, exist := orderMap[id]
		if !exist {
			order = &Order{
				ID:         id,
				AccountID:  accountID,
				CreatedAt:  createdAt,
				TotalPrice: totalPrice,
				Products:   []OrderedProduct{},
			}
			orderMap[id] = order

		}
		order.Products = append(order.Products, OrderedProduct{
			ID:       productID,
			Quantity: quantity,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	var orders []*Order
	for _, order := range orderMap {
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *postgresRepository) ListOrders(ctx context.Context, skip uint64, take uint64) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
		o.id, 
		o.account_id, 
		o.created_at, 
		o.total_price::money::numeric::float8
		op.product_id, 
		op.quantity
		FROM orders o
		INNER JOIN order_products op ON o.id = op.order_id
		ORDER BY o.id DESC
		`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*Order)

	for rows.Next() {
		var (
			id, accountID string
			createdAt     time.Time
			totalPrice    float64
			productID     string
			quantity      uint64
		)
		if err = rows.Scan(&id, &accountID, &createdAt, &totalPrice, &productID, &quantity); err != nil {
			return nil, err
		}

		order, exist := orderMap[id]
		if !exist {
			order = &Order{
				ID:         id,
				AccountID:  accountID,
				CreatedAt:  createdAt,
				TotalPrice: totalPrice,
				Products:   []OrderedProduct{},
			}
			orderMap[id] = order
		}
		order.Products = append(order.Products, OrderedProduct{
			ID:       productID,
			Quantity: quantity,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	var orders []Order
	for _, order := range orderMap {
		orders = append(orders, *order)
	}
	return orders, nil
}
