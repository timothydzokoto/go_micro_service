package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"gopkg.in/olivere/elastic.v6"
)

var (
	ErrNotFound = errors.New("Entity not found")
)

type Repository interface {
	Close()
	PutProduct(ctx context.Context, p Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticsearchRepository struct {
	client *elastic.Client
}

type ProductDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticsearchRepository(url string) (Repository, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, err
	}

	return &elasticsearchRepository{client}, nil
}

func (r *elasticsearchRepository) Close() {
	r.client.Stop()
}

func (r *elasticsearchRepository) PutProduct(ctx context.Context, p Product) error {
	_, err := r.client.Index().
		Index("catalog").
		Type("product").
		Id(p.ID).
		BodyJson(ProductDocument{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}).
		Do(ctx)

	return err
}

func (r *elasticsearchRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get().
		Index("catalog").
		Type("product").
		Id(id).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, ErrNotFound
	}

	var doc ProductDocument
	if err = json.Unmarshal(*res.Source, &doc); err != nil {
		return nil, err
	}

	return &Product{
		ID:          id,
		Name:        doc.Name,
		Description: doc.Description,
		Price:       doc.Price,
	}, nil
}

func (r *elasticsearchRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		Index("catalog").
		Type("product").
		Query(elastic.NewMatchAllQuery()).
		From(int(skip)).
		Size(int(take)).
		Do(ctx)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	products := []Product{}

	for _, hit := range res.Hits.Hits {
		p := ProductDocument{}
		if err = json.Unmarshal(*hit.Source, &p); err != nil {
			return nil, err
		}
		products = append(products, Product{
			ID:          hit.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}

	return products, err
}

func (r *elasticsearchRepository) ListProductWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	items := []*elastic.MultiGetItem{}

	var version int64 = 6 // Or get the version from elsewhere
	versionPtr := &version

	for _, id := range ids {
		items = append(items, elastic.NewMultiGetItem().
			Index("catalog").
			Type("product").
			Id(id).
			Version(*versionPtr).
			VersionType("external"))
	}
	res, err := r.client.Mget().
		Add(items...).
		Do(ctx)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	products := []Product{}

	for _, doc := range res.Docs {
		p := ProductDocument{}
		if err = json.Unmarshal(*doc.Source, &p); err != nil {
			return nil, err
		}
		products = append(products, Product{
			ID:          doc.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}

	return products, err
}

func (r *elasticsearchRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		Index("catalog").
		Type("product").
		Query(elastic.NewMultiMatchQuery(query)).
		From(int(skip)).
		Size(int(take)).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	products := []Product{}

	for _, hit := range res.Hits.Hits {
		p := ProductDocument{}
		if err = json.Unmarshal(*hit.Source, &p); err != nil {
			return nil, err
		}
		products = append(products, Product{
			ID:          hit.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}

	return products, nil
}
