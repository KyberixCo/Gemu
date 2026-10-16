package library

import (
	"database/sql"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"kyberix.com/gemu/internal/storage"
)

type SearchRequest struct {
	Query         string `json:"query"`
	SourceID      string `json:"sourceId"`
	FavoritesOnly bool   `json:"favoritesOnly"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
}

type SearchHit struct {
	GameID   string  `json:"gameId"`
	Title    string  `json:"title"`
	SourceID string  `json:"sourceId"`
	Kind     string  `json:"kind"`
	Score    float64 `json:"score"`
}

type SearchPage struct {
	Hits    []SearchHit `json:"hits"`
	Total   int         `json:"total"`
	HasMore bool        `json:"hasMore"`
}

// Build a small literal language: words are prefixes, quoted text is a phrase,
// and all clauses must match. No user-supplied FTS operators reach MATCH.
func searchExpression(query string) (string, error) {
	if !utf8.ValidString(query) || len(query) > 2048 {
		return "", fmt.Errorf("search query is too long or invalid")
	}
	var clauses []string
	var word strings.Builder
	var phrase []string
	quoted := false
	flushWord := func() {
		if word.Len() == 0 {
			return
		}
		value := word.String()
		word.Reset()
		if quoted {
			phrase = append(phrase, value)
		} else {
			clauses = append(clauses, `"`+value+`"*`)
		}
	}
	flushPhrase := func() {
		if len(phrase) > 0 {
			clauses = append(clauses, `"`+strings.Join(phrase, " ")+`"`)
			phrase = nil
		}
	}
	for _, r := range query {
		if r == '"' {
			flushWord()
			if quoted {
				flushPhrase()
			}
			quoted = !quoted
			continue
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) {
			word.WriteRune(r)
		} else {
			flushWord()
		}
	}
	flushWord()
	flushPhrase()
	if len(clauses) > 32 {
		return "", fmt.Errorf("search query has too many terms")
	}
	return strings.Join(clauses, " AND "), nil
}

func (s *Service) Search(request SearchRequest) (SearchPage, error) {
	page := SearchPage{Hits: []SearchHit{}}
	expression, err := searchExpression(request.Query)
	if err != nil {
		return page, err
	}
	if request.Limit == 0 {
		request.Limit = 48
	}
	if request.Limit < 1 || request.Limit > 100 || request.Offset < 0 || request.Offset > 100000 {
		return page, fmt.Errorf("invalid search pagination")
	}
	if expression == "" {
		return page, nil
	}
	db, err := s.database()
	if err != nil {
		return page, err
	}
	defer db.Close()
	// Scope, source and favorites are applied before ranking and pagination.
	const from = ` FROM library_documents_fts
 JOIN library_documents d ON d.id=library_documents_fts.rowid
 WHERE library_documents_fts MATCH ? AND d.scope=?
 AND (?='' OR d.source_id=?)
 AND (?=0 OR EXISTS (
 SELECT 1 FROM preferences p,json_each(p.value) f
 WHERE p.scope=d.scope AND p.key='gemu.favorites' AND f.value=d.game_id))`
	err = storage.Update(db, func(tx *sql.Tx) error {
		args := []any{expression, s.scope, request.SourceID, request.SourceID, request.FavoritesOnly}
		if err := tx.QueryRow("SELECT count(*)"+from, args...).Scan(&page.Total); err != nil {
			return err
		}
		rows, err := tx.Query("SELECT d.game_id,d.title,d.source_id,d.kind,bm25(library_documents_fts,8.0,1.0,3.0,1.0)"+from+" ORDER BY 5,d.title COLLATE NOCASE,d.game_id LIMIT ? OFFSET ?", append(args, request.Limit, request.Offset)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var hit SearchHit
			if err = rows.Scan(&hit.GameID, &hit.Title, &hit.SourceID, &hit.Kind, &hit.Score); err != nil {
				return err
			}
			page.Hits = append(page.Hits, hit)
		}
		page.HasMore = request.Offset+len(page.Hits) < page.Total
		return rows.Err()
	})
	return page, err
}
