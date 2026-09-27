package bot

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

var ErrRetrieveChunks = errors.New("bot: failed to retrieve chunks")

type RetrievedChunk struct {
	Text       string
	Source     string
	Page       int
	Similarity float64
}

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) RetrieveChunks(ctx context.Context, embedding []float32, topK int) ([]RetrievedChunk, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT text, source, page, 1 - (embedding <=> $1) AS similarity
		FROM document_chunks
		ORDER BY embedding <=> $1
		LIMIT $2
	`, pgvector.NewVector(embedding), topK)
	if err != nil {
		return nil, errors.Join(ErrRetrieveChunks, err)
	}
	defer rows.Close()

	var chunks []RetrievedChunk
	for rows.Next() {
		var c RetrievedChunk
		if err := rows.Scan(&c.Text, &c.Source, &c.Page, &c.Similarity); err != nil {
			return nil, errors.Join(ErrRetrieveChunks, err)
		}
		chunks = append(chunks, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(ErrRetrieveChunks, err)
	}

	return chunks, nil
}
